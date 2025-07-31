// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/access/AccessControl.sol";

/**
 * @title WrappedToken
 * @dev ERC20 token representing locked tokens from another chain
 * @notice This contract creates wrapped representations of tokens locked on other chains
 */
contract WrappedToken is ERC20, AccessControl {
    bytes32 public constant MINTER_ROLE = keccak256("MINTER_ROLE");
    bytes32 public constant BURNER_ROLE = keccak256("BURNER_ROLE");

    // Original token information
    address public immutable originalToken;
    uint256 public immutable originalChainId;
    uint8 private _decimals;

    // Events
    event TokensMinted(address indexed to, uint256 amount);
    event TokensBurned(address indexed from, uint256 amount);

    // Custom errors
    error ZeroAddress();
    error ZeroAmount();
    error InsufficientBalance(uint256 requested, uint256 available);

    /**
     * @dev Constructor for wrapped token
     * @param name Name of the wrapped token
     * @param symbol Symbol of the wrapped token
     * @param decimals_ Number of decimals for the token
     * @param originalToken_ Address of the original token on source chain
     * @param originalChainId_ Chain ID where the original token exists
     * @param bridge Address of the bridge contract that can mint/burn
     */
    constructor(
        string memory name,
        string memory symbol,
        uint8 decimals_,
        address originalToken_,
        uint256 originalChainId_,
        address bridge
    ) ERC20(name, symbol) {
        if (originalToken_ == address(0)) revert ZeroAddress();
        if (bridge == address(0)) revert ZeroAddress();

        _decimals = decimals_;
        originalToken = originalToken_;
        originalChainId = originalChainId_;

        _grantRole(DEFAULT_ADMIN_ROLE, bridge);
        _grantRole(MINTER_ROLE, bridge);
        _grantRole(BURNER_ROLE, bridge);
    }

    /**
     * @dev Returns the number of decimals used to get its user representation
     */
    function decimals() public view virtual override returns (uint8) {
        return _decimals;
    }

    /**
     * @dev Mints tokens to specified address
     * @param to Address to mint tokens to
     * @param amount Amount of tokens to mint
     */
    function mint(address to, uint256 amount) external onlyRole(MINTER_ROLE) {
        if (to == address(0)) revert ZeroAddress();
        if (amount == 0) revert ZeroAmount();

        _mint(to, amount);
        emit TokensMinted(to, amount);
    }

    /**
     * @dev Burns tokens from specified address
     * @param from Address to burn tokens from
     * @param amount Amount of tokens to burn
     */
    function burn(address from, uint256 amount) external onlyRole(BURNER_ROLE) {
        if (from == address(0)) revert ZeroAddress();
        if (amount == 0) revert ZeroAmount();
        if (balanceOf(from) < amount) {
            revert InsufficientBalance(amount, balanceOf(from));
        }

        _burn(from, amount);
        emit TokensBurned(from, amount);
    }

    /**
     * @dev Burns tokens from caller's balance
     * @param amount Amount of tokens to burn
     */
    function burn(uint256 amount) external {
        if (amount == 0) revert ZeroAmount();
        if (balanceOf(msg.sender) < amount) {
            revert InsufficientBalance(amount, balanceOf(msg.sender));
        }

        _burn(msg.sender, amount);
        emit TokensBurned(msg.sender, amount);
    }
}