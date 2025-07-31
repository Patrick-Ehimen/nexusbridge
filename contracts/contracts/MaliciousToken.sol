// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";

/**
 * @title MaliciousToken
 * @dev Token contract that attempts reentrancy attacks for testing purposes
 */
contract MaliciousToken is ERC20 {
    address public bridge;
    bool public attacking = false;

    constructor(address _bridge) ERC20("Malicious", "MAL") {
        bridge = _bridge;
        _mint(address(this), 1000 ether);
    }

    function triggerReentrancy() external {
        attacking = true;
        this.approve(bridge, 100 ether);

        // This will trigger the transfer hook and attempt reentrancy
        (bool success, ) = bridge.call(
            abi.encodeWithSignature(
                "lockTokens(address,uint256,uint256,address)",
                address(this),
                100 ether,
                137,
                msg.sender
            )
        );
        require(success, "Initial call failed");
    }

    function _update(
        address from,
        address to,
        uint256 value
    ) internal override {
        super._update(from, to, value);

        if (attacking && to == bridge) {
            attacking = false;
            // Attempt reentrancy
            (bool success, ) = bridge.call(
                abi.encodeWithSignature(
                    "lockTokens(address,uint256,uint256,address)",
                    address(this),
                    50 ether,
                    137,
                    msg.sender
                )
            );
            // This should fail due to reentrancy guard
        }
    }
}
