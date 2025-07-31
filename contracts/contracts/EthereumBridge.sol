// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/access/AccessControl.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "@openzeppelin/contracts/utils/Pausable.sol";
import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import "@openzeppelin/contracts/utils/cryptography/MessageHashUtils.sol";

/**
 * @title EthereumBridge
 * @dev Smart contract for locking/unlocking tokens on Ethereum for cross-chain transfers
 * @notice This contract handles the Ethereum side of the NexusBridge protocol
 */
contract EthereumBridge is AccessControl, ReentrancyGuard, Pausable {
    using SafeERC20 for IERC20;
    using ECDSA for bytes32;
    using MessageHashUtils for bytes32;

    // Role definitions
    bytes32 public constant RELAYER_ROLE = keccak256("RELAYER_ROLE");
    bytes32 public constant ADMIN_ROLE = keccak256("ADMIN_ROLE");

    // Struct to store transfer information
    struct Transfer {
        address token;
        uint256 amount;
        address sender;
        address recipient;
        uint256 destinationChain;
        bool completed;
        uint256 timestamp;
    }

    // State variables
    mapping(bytes32 => Transfer) public transfers;
    mapping(bytes32 => bool) public processedTransfers;
    mapping(address => bool) public supportedTokens;
    mapping(address => uint256) public lockedBalances;

    address[] public relayers;
    uint256 public requiredSignatures;
    uint256 public transferNonce;

    // Events
    event TokensLocked(
        bytes32 indexed transferId,
        address indexed user,
        address indexed token,
        uint256 amount,
        uint256 destinationChain,
        address recipient,
        uint256 timestamp
    );

    event TokensUnlocked(
        bytes32 indexed transferId,
        address indexed recipient,
        address indexed token,
        uint256 amount,
        uint256 timestamp
    );

    event TokenSupported(address indexed token, bool supported);
    event RelayerAdded(address indexed relayer);
    event RelayerRemoved(address indexed relayer);
    event RequiredSignaturesUpdated(uint256 newRequiredSignatures);

    // Custom errors
    error TokenNotSupported(address token);
    error InsufficientAmount();
    error TransferAlreadyProcessed(bytes32 transferId);
    error InvalidSignatures();
    error InsufficientSignatures(uint256 provided, uint256 required);
    error InvalidTransfer(bytes32 transferId);
    error InsufficientLockedBalance(
        address token,
        uint256 requested,
        uint256 available
    );
    error InvalidDestinationChain(uint256 chainId);
    error ZeroAddress();
    error InvalidSignatureLength();

    /**
     * @dev Constructor sets up initial roles and configuration
     * @param _admin Address that will have admin role
     * @param _relayers Array of initial relayer addresses
     * @param _requiredSignatures Number of signatures required for unlock operations
     */
    constructor(
        address _admin,
        address[] memory _relayers,
        uint256 _requiredSignatures
    ) {
        if (_admin == address(0)) revert ZeroAddress();
        if (
            _requiredSignatures == 0 || _requiredSignatures > _relayers.length
        ) {
            revert InsufficientSignatures(
                _requiredSignatures,
                _relayers.length
            );
        }

        _grantRole(DEFAULT_ADMIN_ROLE, _admin);
        _grantRole(ADMIN_ROLE, _admin);

        // Set up relayers
        for (uint256 i = 0; i < _relayers.length; i++) {
            if (_relayers[i] == address(0)) revert ZeroAddress();
            _grantRole(RELAYER_ROLE, _relayers[i]);
            relayers.push(_relayers[i]);
            emit RelayerAdded(_relayers[i]);
        }

        requiredSignatures = _requiredSignatures;
        emit RequiredSignaturesUpdated(_requiredSignatures);
    }

    /**
     * @dev Locks tokens for cross-chain transfer
     * @param token Address of the ERC20 token to lock
     * @param amount Amount of tokens to lock
     * @param destinationChain Chain ID of the destination chain
     * @param recipient Address that will receive tokens on destination chain
     * @return transferId Unique identifier for this transfer
     */
    function lockTokens(
        address token,
        uint256 amount,
        uint256 destinationChain,
        address recipient
    ) external nonReentrant whenNotPaused returns (bytes32 transferId) {
        if (!supportedTokens[token]) revert TokenNotSupported(token);
        if (amount == 0) revert InsufficientAmount();
        if (recipient == address(0)) revert ZeroAddress();
        if (destinationChain == block.chainid)
            revert InvalidDestinationChain(destinationChain);

        // Generate unique transfer ID
        transferId = keccak256(
            abi.encodePacked(
                msg.sender,
                token,
                amount,
                destinationChain,
                recipient,
                transferNonce++,
                block.timestamp
            )
        );

        // Store transfer information
        transfers[transferId] = Transfer({
            token: token,
            amount: amount,
            sender: msg.sender,
            recipient: recipient,
            destinationChain: destinationChain,
            completed: false,
            timestamp: block.timestamp
        });

        // Transfer tokens from user to contract
        IERC20(token).safeTransferFrom(msg.sender, address(this), amount);

        // Update locked balance
        lockedBalances[token] += amount;

        emit TokensLocked(
            transferId,
            msg.sender,
            token,
            amount,
            destinationChain,
            recipient,
            block.timestamp
        );

        return transferId;
    }

    /**
     * @dev Unlocks tokens with multi-signature validation
     * @param transferId Unique identifier for the transfer
     * @param token Address of the ERC20 token to unlock
     * @param amount Amount of tokens to unlock
     * @param recipient Address that will receive the unlocked tokens
     * @param signatures Array of signatures from relayers
     */
    function unlockTokens(
        bytes32 transferId,
        address token,
        uint256 amount,
        address recipient,
        bytes[] calldata signatures
    ) external nonReentrant whenNotPaused {
        if (processedTransfers[transferId])
            revert TransferAlreadyProcessed(transferId);
        if (!supportedTokens[token]) revert TokenNotSupported(token);
        if (amount == 0) revert InsufficientAmount();
        if (recipient == address(0)) revert ZeroAddress();
        if (signatures.length < requiredSignatures) {
            revert InsufficientSignatures(
                signatures.length,
                requiredSignatures
            );
        }
        if (lockedBalances[token] < amount) {
            revert InsufficientLockedBalance(
                token,
                amount,
                lockedBalances[token]
            );
        }

        // Verify signatures
        _verifySignatures(transferId, token, amount, recipient, signatures);

        // Mark transfer as processed
        processedTransfers[transferId] = true;

        // Update locked balance
        lockedBalances[token] -= amount;

        // Transfer tokens to recipient
        IERC20(token).safeTransfer(recipient, amount);

        emit TokensUnlocked(
            transferId,
            recipient,
            token,
            amount,
            block.timestamp
        );
    }

    /**
     * @dev Verifies multi-signatures for unlock operation
     * @param transferId Unique identifier for the transfer
     * @param token Address of the ERC20 token
     * @param amount Amount of tokens
     * @param recipient Address that will receive tokens
     * @param signatures Array of signatures to verify
     */
    function _verifySignatures(
        bytes32 transferId,
        address token,
        uint256 amount,
        address recipient,
        bytes[] calldata signatures
    ) internal view {
        // Create message hash
        bytes32 messageHash = keccak256(
            abi.encodePacked(
                transferId,
                token,
                amount,
                recipient,
                block.chainid
            )
        );
        bytes32 ethSignedMessageHash = messageHash.toEthSignedMessageHash();

        address[] memory signers = new address[](signatures.length);
        uint256 validSignatures = 0;

        // Verify each signature
        for (uint256 i = 0; i < signatures.length; i++) {
            if (signatures[i].length != 65) revert InvalidSignatureLength();

            address signer = ethSignedMessageHash.recover(signatures[i]);

            // Check if signer is a relayer and hasn't signed already
            if (hasRole(RELAYER_ROLE, signer)) {
                bool alreadySigned = false;
                for (uint256 j = 0; j < validSignatures; j++) {
                    if (signers[j] == signer) {
                        alreadySigned = true;
                        break;
                    }
                }

                if (!alreadySigned) {
                    signers[validSignatures] = signer;
                    validSignatures++;
                }
            }
        }

        if (validSignatures < requiredSignatures) {
            revert InsufficientSignatures(validSignatures, requiredSignatures);
        }
    }

    /**
     * @dev Adds support for a new token
     * @param token Address of the ERC20 token to support
     */
    function addSupportedToken(address token) external onlyRole(ADMIN_ROLE) {
        if (token == address(0)) revert ZeroAddress();
        supportedTokens[token] = true;
        emit TokenSupported(token, true);
    }

    /**
     * @dev Removes support for a token
     * @param token Address of the ERC20 token to remove support for
     */
    function removeSupportedToken(address token) external onlyRole(ADMIN_ROLE) {
        supportedTokens[token] = false;
        emit TokenSupported(token, false);
    }

    /**
     * @dev Adds a new relayer
     * @param relayer Address of the new relayer
     */
    function addRelayer(address relayer) external onlyRole(ADMIN_ROLE) {
        if (relayer == address(0)) revert ZeroAddress();
        if (hasRole(RELAYER_ROLE, relayer)) return; // Already a relayer

        _grantRole(RELAYER_ROLE, relayer);
        relayers.push(relayer);
        emit RelayerAdded(relayer);
    }

    /**
     * @dev Removes a relayer
     * @param relayer Address of the relayer to remove
     */
    function removeRelayer(address relayer) external onlyRole(ADMIN_ROLE) {
        if (!hasRole(RELAYER_ROLE, relayer)) return; // Not a relayer

        _revokeRole(RELAYER_ROLE, relayer);

        // Remove from relayers array
        for (uint256 i = 0; i < relayers.length; i++) {
            if (relayers[i] == relayer) {
                relayers[i] = relayers[relayers.length - 1];
                relayers.pop();
                break;
            }
        }

        emit RelayerRemoved(relayer);
    }

    /**
     * @dev Updates the required number of signatures
     * @param _requiredSignatures New required signature count
     */
    function updateRequiredSignatures(
        uint256 _requiredSignatures
    ) external onlyRole(ADMIN_ROLE) {
        if (_requiredSignatures == 0 || _requiredSignatures > relayers.length) {
            revert InsufficientSignatures(_requiredSignatures, relayers.length);
        }

        requiredSignatures = _requiredSignatures;
        emit RequiredSignaturesUpdated(_requiredSignatures);
    }

    /**
     * @dev Pauses the contract (emergency stop)
     */
    function pause() external onlyRole(ADMIN_ROLE) {
        _pause();
    }

    /**
     * @dev Unpauses the contract
     */
    function unpause() external onlyRole(ADMIN_ROLE) {
        _unpause();
    }

    /**
     * @dev Gets transfer information
     * @param transferId Unique identifier for the transfer
     * @return Transfer struct containing transfer details
     */
    function getTransfer(
        bytes32 transferId
    ) external view returns (Transfer memory) {
        return transfers[transferId];
    }

    /**
     * @dev Checks if a transfer has been processed
     * @param transferId Unique identifier for the transfer
     * @return bool indicating if transfer is processed
     */
    function isTransferProcessed(
        bytes32 transferId
    ) external view returns (bool) {
        return processedTransfers[transferId];
    }

    /**
     * @dev Gets the list of relayers
     * @return Array of relayer addresses
     */
    function getRelayers() external view returns (address[] memory) {
        return relayers;
    }

    /**
     * @dev Gets the locked balance for a token
     * @param token Address of the ERC20 token
     * @return Amount of tokens locked in the contract
     */
    function getLockedBalance(address token) external view returns (uint256) {
        return lockedBalances[token];
    }

    /**
     * @dev Checks if a token is supported
     * @param token Address of the ERC20 token
     * @return bool indicating if token is supported
     */
    function isTokenSupported(address token) external view returns (bool) {
        return supportedTokens[token];
    }

    /**
     * @dev Emergency function to recover stuck tokens (only admin)
     * @param token Address of the ERC20 token to recover
     * @param amount Amount of tokens to recover
     * @param to Address to send recovered tokens to
     */
    function emergencyRecoverTokens(
        address token,
        uint256 amount,
        address to
    ) external onlyRole(DEFAULT_ADMIN_ROLE) {
        if (to == address(0)) revert ZeroAddress();
        IERC20(token).safeTransfer(to, amount);
    }
}
