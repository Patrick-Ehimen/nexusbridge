// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/access/AccessControl.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "@openzeppelin/contracts/utils/Pausable.sol";
import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import "@openzeppelin/contracts/utils/cryptography/MessageHashUtils.sol";
import "./WrappedToken.sol";

/**
 * @title PolygonBridge
 * @dev Smart contract for minting/burning wrapped tokens on Polygon for cross-chain transfers
 * @notice This contract handles the Polygon side of the NexusBridge protocol
 */
contract PolygonBridge is AccessControl, ReentrancyGuard, Pausable {
    using ECDSA for bytes32;
    using MessageHashUtils for bytes32;

    // Role definitions
    bytes32 public constant RELAYER_ROLE = keccak256("RELAYER_ROLE");
    bytes32 public constant ADMIN_ROLE = keccak256("ADMIN_ROLE");

    // Struct to store supported token information
    struct TokenInfo {
        address wrappedToken;
        address originalToken;
        uint256 originalChainId;
        string name;
        string symbol;
        uint8 decimals;
        bool isSupported;
        uint256 totalMinted;
    }

    // Struct to store transfer information
    struct Transfer {
        address originalToken;
        address wrappedToken;
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
    mapping(bytes32 => TokenInfo) public supportedTokens; // keccak256(originalToken, originalChainId) => TokenInfo
    mapping(address => bytes32) public wrappedToOriginal; // wrapped token address => original token key

    address[] public relayers;
    uint256 public requiredSignatures;
    uint256 public transferNonce;

    // Events
    event TokensMinted(
        bytes32 indexed transferId,
        address indexed recipient,
        address indexed wrappedToken,
        address originalToken,
        uint256 originalChainId,
        uint256 amount,
        uint256 timestamp
    );

    event TokensBurned(
        bytes32 indexed transferId,
        address indexed user,
        address indexed wrappedToken,
        address originalToken,
        uint256 amount,
        uint256 destinationChain,
        address recipient,
        uint256 timestamp
    );

    event WrappedTokenDeployed(
        bytes32 indexed tokenKey,
        address indexed wrappedToken,
        address indexed originalToken,
        uint256 originalChainId,
        string name,
        string symbol,
        uint8 decimals
    );

    event TokenSupported(
        bytes32 indexed tokenKey,
        address originalToken,
        uint256 originalChainId,
        bool supported
    );

    event RelayerAdded(address indexed relayer);
    event RelayerRemoved(address indexed relayer);
    event RequiredSignaturesUpdated(uint256 newRequiredSignatures);

    // Custom errors
    error TokenNotSupported(address originalToken, uint256 originalChainId);
    error TokenAlreadySupported(address originalToken, uint256 originalChainId);
    error WrappedTokenNotFound(address wrappedToken);
    error InsufficientAmount();
    error TransferAlreadyProcessed(bytes32 transferId);
    error InvalidSignatures();
    error InsufficientSignatures(uint256 provided, uint256 required);
    error InvalidTransfer(bytes32 transferId);
    error InvalidDestinationChain(uint256 chainId);
    error ZeroAddress();
    error InvalidSignatureLength();
    error WrappedTokenDeploymentFailed();

    /**
     * @dev Constructor sets up initial roles and configuration
     * @param _admin Address that will have admin role
     * @param _relayers Array of initial relayer addresses
     * @param _requiredSignatures Number of signatures required for mint operations
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
     * @dev Deploys a new wrapped token contract
     * @param originalToken Address of the original token on source chain
     * @param originalChainId Chain ID where the original token exists
     * @param name Name for the wrapped token
     * @param symbol Symbol for the wrapped token
     * @param decimals Number of decimals for the wrapped token
     * @return wrappedToken Address of the deployed wrapped token
     */
    function deployWrappedToken(
        address originalToken,
        uint256 originalChainId,
        string memory name,
        string memory symbol,
        uint8 decimals
    ) external onlyRole(ADMIN_ROLE) returns (address wrappedToken) {
        if (originalToken == address(0)) revert ZeroAddress();
        if (originalChainId == block.chainid)
            revert InvalidDestinationChain(originalChainId);

        bytes32 tokenKey = keccak256(
            abi.encodePacked(originalToken, originalChainId)
        );

        if (supportedTokens[tokenKey].isSupported) {
            revert TokenAlreadySupported(originalToken, originalChainId);
        }

        // Deploy wrapped token
        try
            new WrappedToken(
                name,
                symbol,
                decimals,
                originalToken,
                originalChainId,
                address(this)
            )
        returns (WrappedToken token) {
            wrappedToken = address(token);
        } catch {
            revert WrappedTokenDeploymentFailed();
        }

        // Store token information
        supportedTokens[tokenKey] = TokenInfo({
            wrappedToken: wrappedToken,
            originalToken: originalToken,
            originalChainId: originalChainId,
            name: name,
            symbol: symbol,
            decimals: decimals,
            isSupported: true,
            totalMinted: 0
        });

        wrappedToOriginal[wrappedToken] = tokenKey;

        emit WrappedTokenDeployed(
            tokenKey,
            wrappedToken,
            originalToken,
            originalChainId,
            name,
            symbol,
            decimals
        );

        emit TokenSupported(tokenKey, originalToken, originalChainId, true);

        return wrappedToken;
    }

    /**
     * @dev Mints wrapped tokens with multi-signature validation
     * @param transferId Unique identifier for the transfer
     * @param originalToken Address of the original token on source chain
     * @param originalChainId Chain ID where the original token exists
     * @param amount Amount of tokens to mint
     * @param recipient Address that will receive the minted tokens
     * @param signatures Array of signatures from relayers
     */
    function mintTokens(
        bytes32 transferId,
        address originalToken,
        uint256 originalChainId,
        uint256 amount,
        address recipient,
        bytes[] calldata signatures
    ) external nonReentrant whenNotPaused {
        if (processedTransfers[transferId])
            revert TransferAlreadyProcessed(transferId);
        if (amount == 0) revert InsufficientAmount();
        if (recipient == address(0)) revert ZeroAddress();
        if (signatures.length < requiredSignatures) {
            revert InsufficientSignatures(
                signatures.length,
                requiredSignatures
            );
        }

        bytes32 tokenKey = keccak256(
            abi.encodePacked(originalToken, originalChainId)
        );

        if (!supportedTokens[tokenKey].isSupported) {
            revert TokenNotSupported(originalToken, originalChainId);
        }

        // Verify signatures
        _verifySignatures(
            transferId,
            originalToken,
            originalChainId,
            amount,
            recipient,
            signatures
        );

        // Mark transfer as processed
        processedTransfers[transferId] = true;

        TokenInfo storage tokenInfo = supportedTokens[tokenKey];
        address wrappedToken = tokenInfo.wrappedToken;

        // Update total minted
        tokenInfo.totalMinted += amount;

        // Mint wrapped tokens
        WrappedToken(wrappedToken).mint(recipient, amount);

        emit TokensMinted(
            transferId,
            recipient,
            wrappedToken,
            originalToken,
            originalChainId,
            amount,
            block.timestamp
        );
    }

    /**
     * @dev Burns wrapped tokens for cross-chain transfer
     * @param wrappedToken Address of the wrapped token to burn
     * @param amount Amount of tokens to burn
     * @param destinationChain Chain ID of the destination chain
     * @param recipient Address that will receive tokens on destination chain
     * @return transferId Unique identifier for this transfer
     */
    function burnTokens(
        address wrappedToken,
        uint256 amount,
        uint256 destinationChain,
        address recipient
    ) external nonReentrant whenNotPaused returns (bytes32 transferId) {
        if (amount == 0) revert InsufficientAmount();
        if (recipient == address(0)) revert ZeroAddress();
        if (destinationChain == block.chainid)
            revert InvalidDestinationChain(destinationChain);

        bytes32 tokenKey = wrappedToOriginal[wrappedToken];
        if (tokenKey == bytes32(0)) {
            revert WrappedTokenNotFound(wrappedToken);
        }

        TokenInfo storage tokenInfo = supportedTokens[tokenKey];
        if (!tokenInfo.isSupported) {
            revert TokenNotSupported(
                tokenInfo.originalToken,
                tokenInfo.originalChainId
            );
        }

        // Generate unique transfer ID
        transferId = keccak256(
            abi.encodePacked(
                msg.sender,
                wrappedToken,
                amount,
                destinationChain,
                recipient,
                transferNonce++,
                block.timestamp
            )
        );

        // Store transfer information
        transfers[transferId] = Transfer({
            originalToken: tokenInfo.originalToken,
            wrappedToken: wrappedToken,
            amount: amount,
            sender: msg.sender,
            recipient: recipient,
            destinationChain: destinationChain,
            completed: false,
            timestamp: block.timestamp
        });

        // Update total minted (decrease)
        tokenInfo.totalMinted -= amount;

        // Burn wrapped tokens
        WrappedToken(wrappedToken).burn(msg.sender, amount);

        emit TokensBurned(
            transferId,
            msg.sender,
            wrappedToken,
            tokenInfo.originalToken,
            amount,
            destinationChain,
            recipient,
            block.timestamp
        );

        return transferId;
    }

    /**
     * @dev Verifies multi-signatures for mint operation
     * @param transferId Unique identifier for the transfer
     * @param originalToken Address of the original token
     * @param originalChainId Chain ID where the original token exists
     * @param amount Amount of tokens
     * @param recipient Address that will receive tokens
     * @param signatures Array of signatures to verify
     */
    function _verifySignatures(
        bytes32 transferId,
        address originalToken,
        uint256 originalChainId,
        uint256 amount,
        address recipient,
        bytes[] calldata signatures
    ) internal view {
        // Create message hash
        bytes32 messageHash = keccak256(
            abi.encodePacked(
                transferId,
                originalToken,
                originalChainId,
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
     * @dev Adds support for an existing wrapped token
     * @param originalToken Address of the original token on source chain
     * @param originalChainId Chain ID where the original token exists
     * @param wrappedToken Address of the existing wrapped token
     */
    function addSupportedToken(
        address originalToken,
        uint256 originalChainId,
        address wrappedToken
    ) external onlyRole(ADMIN_ROLE) {
        if (originalToken == address(0)) revert ZeroAddress();
        if (wrappedToken == address(0)) revert ZeroAddress();

        bytes32 tokenKey = keccak256(
            abi.encodePacked(originalToken, originalChainId)
        );

        if (supportedTokens[tokenKey].isSupported) {
            revert TokenAlreadySupported(originalToken, originalChainId);
        }

        // Get token details from wrapped token contract
        WrappedToken token = WrappedToken(wrappedToken);

        supportedTokens[tokenKey] = TokenInfo({
            wrappedToken: wrappedToken,
            originalToken: originalToken,
            originalChainId: originalChainId,
            name: token.name(),
            symbol: token.symbol(),
            decimals: token.decimals(),
            isSupported: true,
            totalMinted: token.totalSupply()
        });

        wrappedToOriginal[wrappedToken] = tokenKey;

        emit TokenSupported(tokenKey, originalToken, originalChainId, true);
    }

    /**
     * @dev Removes support for a token
     * @param originalToken Address of the original token on source chain
     * @param originalChainId Chain ID where the original token exists
     */
    function removeSupportedToken(
        address originalToken,
        uint256 originalChainId
    ) external onlyRole(ADMIN_ROLE) {
        bytes32 tokenKey = keccak256(
            abi.encodePacked(originalToken, originalChainId)
        );

        TokenInfo storage tokenInfo = supportedTokens[tokenKey];
        if (!tokenInfo.isSupported) {
            revert TokenNotSupported(originalToken, originalChainId);
        }

        tokenInfo.isSupported = false;
        delete wrappedToOriginal[tokenInfo.wrappedToken];

        emit TokenSupported(tokenKey, originalToken, originalChainId, false);
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
     * @dev Gets token information
     * @param originalToken Address of the original token on source chain
     * @param originalChainId Chain ID where the original token exists
     * @return TokenInfo struct containing token details
     */
    function getTokenInfo(
        address originalToken,
        uint256 originalChainId
    ) external view returns (TokenInfo memory) {
        bytes32 tokenKey = keccak256(
            abi.encodePacked(originalToken, originalChainId)
        );
        return supportedTokens[tokenKey];
    }

    /**
     * @dev Gets wrapped token address for original token
     * @param originalToken Address of the original token on source chain
     * @param originalChainId Chain ID where the original token exists
     * @return address of the wrapped token
     */
    function getWrappedToken(
        address originalToken,
        uint256 originalChainId
    ) external view returns (address) {
        bytes32 tokenKey = keccak256(
            abi.encodePacked(originalToken, originalChainId)
        );
        return supportedTokens[tokenKey].wrappedToken;
    }

    /**
     * @dev Gets original token information for wrapped token
     * @param wrappedToken Address of the wrapped token
     * @return originalToken Address of the original token
     * @return originalChainId Chain ID where the original token exists
     */
    function getOriginalToken(
        address wrappedToken
    ) external view returns (address originalToken, uint256 originalChainId) {
        bytes32 tokenKey = wrappedToOriginal[wrappedToken];
        if (tokenKey == bytes32(0)) {
            revert WrappedTokenNotFound(wrappedToken);
        }

        TokenInfo memory tokenInfo = supportedTokens[tokenKey];
        return (tokenInfo.originalToken, tokenInfo.originalChainId);
    }

    /**
     * @dev Checks if a token is supported
     * @param originalToken Address of the original token on source chain
     * @param originalChainId Chain ID where the original token exists
     * @return bool indicating if token is supported
     */
    function isTokenSupported(
        address originalToken,
        uint256 originalChainId
    ) external view returns (bool) {
        bytes32 tokenKey = keccak256(
            abi.encodePacked(originalToken, originalChainId)
        );
        return supportedTokens[tokenKey].isSupported;
    }

    /**
     * @dev Gets the list of relayers
     * @return Array of relayer addresses
     */
    function getRelayers() external view returns (address[] memory) {
        return relayers;
    }

    /**
     * @dev Gets total minted amount for a token
     * @param originalToken Address of the original token on source chain
     * @param originalChainId Chain ID where the original token exists
     * @return Total amount of wrapped tokens minted
     */
    function getTotalMinted(
        address originalToken,
        uint256 originalChainId
    ) external view returns (uint256) {
        bytes32 tokenKey = keccak256(
            abi.encodePacked(originalToken, originalChainId)
        );
        return supportedTokens[tokenKey].totalMinted;
    }
}
