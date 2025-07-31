const { expect } = require("chai");
const { ethers } = require("hardhat");

describe("WrappedToken", function () {
  let wrappedToken;
  let mockERC20;
  let bridge, user1, user2;

  const ETHEREUM_CHAIN_ID = 1;

  beforeEach(async function () {
    [bridge, user1, user2] = await ethers.getSigners();

    // Deploy MockERC20 for testing
    const MockERC20 = await ethers.getContractFactory("MockERC20");
    mockERC20 = await MockERC20.deploy("Test Token", "TEST", 18);
    await mockERC20.waitForDeployment();

    // Deploy WrappedToken
    const WrappedToken = await ethers.getContractFactory("WrappedToken");
    wrappedToken = await WrappedToken.deploy(
      "Wrapped Test Token",
      "wTEST",
      18,
      mockERC20.target,
      ETHEREUM_CHAIN_ID,
      bridge.address
    );
    await wrappedToken.waitForDeployment();
  });

  describe("Deployment", function () {
    it("Should set the correct token details", async function () {
      expect(await wrappedToken.name()).to.equal("Wrapped Test Token");
      expect(await wrappedToken.symbol()).to.equal("wTEST");
      expect(await wrappedToken.decimals()).to.equal(18);
      expect(await wrappedToken.originalToken()).to.equal(mockERC20.target);
      expect(await wrappedToken.originalChainId()).to.equal(ETHEREUM_CHAIN_ID);
    });

    it("Should grant correct roles to bridge", async function () {
      const MINTER_ROLE = await wrappedToken.MINTER_ROLE();
      const BURNER_ROLE = await wrappedToken.BURNER_ROLE();
      const DEFAULT_ADMIN_ROLE = await wrappedToken.DEFAULT_ADMIN_ROLE();

      expect(await wrappedToken.hasRole(MINTER_ROLE, bridge.address)).to.be
        .true;
      expect(await wrappedToken.hasRole(BURNER_ROLE, bridge.address)).to.be
        .true;
      expect(await wrappedToken.hasRole(DEFAULT_ADMIN_ROLE, bridge.address)).to
        .be.true;
    });

    it("Should revert with zero original token address", async function () {
      const WrappedToken = await ethers.getContractFactory("WrappedToken");
      await expect(
        WrappedToken.deploy(
          "Wrapped Test Token",
          "wTEST",
          18,
          ethers.ZeroAddress,
          ETHEREUM_CHAIN_ID,
          bridge.address
        )
      ).to.be.revertedWithCustomError(wrappedToken, "ZeroAddress");
    });

    it("Should revert with zero bridge address", async function () {
      const WrappedToken = await ethers.getContractFactory("WrappedToken");
      await expect(
        WrappedToken.deploy(
          "Wrapped Test Token",
          "wTEST",
          18,
          mockERC20.target,
          ETHEREUM_CHAIN_ID,
          ethers.ZeroAddress
        )
      ).to.be.revertedWithCustomError(wrappedToken, "ZeroAddress");
    });
  });

  describe("Minting", function () {
    it("Should mint tokens successfully", async function () {
      const amount = ethers.parseEther("100");

      await expect(wrappedToken.connect(bridge).mint(user1.address, amount))
        .to.emit(wrappedToken, "TokensMinted")
        .withArgs(user1.address, amount);

      expect(await wrappedToken.balanceOf(user1.address)).to.equal(amount);
      expect(await wrappedToken.totalSupply()).to.equal(amount);
    });

    it("Should revert when non-minter tries to mint", async function () {
      const amount = ethers.parseEther("100");

      await expect(wrappedToken.connect(user1).mint(user1.address, amount)).to
        .be.reverted;
    });

    it("Should revert with zero address", async function () {
      const amount = ethers.parseEther("100");

      await expect(
        wrappedToken.connect(bridge).mint(ethers.ZeroAddress, amount)
      ).to.be.revertedWithCustomError(wrappedToken, "ZeroAddress");
    });

    it("Should revert with zero amount", async function () {
      await expect(
        wrappedToken.connect(bridge).mint(user1.address, 0)
      ).to.be.revertedWithCustomError(wrappedToken, "ZeroAmount");
    });
  });

  describe("Burning", function () {
    beforeEach(async function () {
      // Mint some tokens first
      const amount = ethers.parseEther("1000");
      await wrappedToken.connect(bridge).mint(user1.address, amount);
    });

    it("Should burn tokens from address successfully", async function () {
      const burnAmount = ethers.parseEther("100");
      const initialBalance = await wrappedToken.balanceOf(user1.address);

      await expect(
        wrappedToken
          .connect(bridge)
          ["burn(address,uint256)"](user1.address, burnAmount)
      )
        .to.emit(wrappedToken, "TokensBurned")
        .withArgs(user1.address, burnAmount);

      expect(await wrappedToken.balanceOf(user1.address)).to.equal(
        initialBalance - burnAmount
      );
      expect(await wrappedToken.totalSupply()).to.equal(
        initialBalance - burnAmount
      );
    });

    it("Should burn tokens from caller successfully", async function () {
      const burnAmount = ethers.parseEther("100");
      const initialBalance = await wrappedToken.balanceOf(user1.address);

      await expect(wrappedToken.connect(user1).burn(burnAmount))
        .to.emit(wrappedToken, "TokensBurned")
        .withArgs(user1.address, burnAmount);

      expect(await wrappedToken.balanceOf(user1.address)).to.equal(
        initialBalance - burnAmount
      );
    });

    it("Should revert when non-burner tries to burn from address", async function () {
      const burnAmount = ethers.parseEther("100");

      await expect(
        wrappedToken
          .connect(user1)
          ["burn(address,uint256)"](user1.address, burnAmount)
      ).to.be.reverted;
    });

    it("Should revert with zero address", async function () {
      const burnAmount = ethers.parseEther("100");

      await expect(
        wrappedToken
          .connect(bridge)
          ["burn(address,uint256)"](ethers.ZeroAddress, burnAmount)
      ).to.be.revertedWithCustomError(wrappedToken, "ZeroAddress");
    });

    it("Should revert with zero amount", async function () {
      await expect(
        wrappedToken.connect(bridge)["burn(address,uint256)"](user1.address, 0)
      ).to.be.revertedWithCustomError(wrappedToken, "ZeroAmount");
    });

    it("Should revert with insufficient balance", async function () {
      const userBalance = await wrappedToken.balanceOf(user1.address);
      const burnAmount = userBalance + ethers.parseEther("1");

      await expect(
        wrappedToken
          .connect(bridge)
          ["burn(address,uint256)"](user1.address, burnAmount)
      )
        .to.be.revertedWithCustomError(wrappedToken, "InsufficientBalance")
        .withArgs(burnAmount, userBalance);
    });

    it("Should revert when caller burns more than balance", async function () {
      const userBalance = await wrappedToken.balanceOf(user1.address);
      const burnAmount = userBalance + ethers.parseEther("1");

      await expect(wrappedToken.connect(user1).burn(burnAmount))
        .to.be.revertedWithCustomError(wrappedToken, "InsufficientBalance")
        .withArgs(burnAmount, userBalance);
    });
  });

  describe("ERC20 Functionality", function () {
    beforeEach(async function () {
      // Mint some tokens first
      const amount = ethers.parseEther("1000");
      await wrappedToken.connect(bridge).mint(user1.address, amount);
    });

    it("Should transfer tokens correctly", async function () {
      const transferAmount = ethers.parseEther("100");

      await expect(
        wrappedToken.connect(user1).transfer(user2.address, transferAmount)
      )
        .to.emit(wrappedToken, "Transfer")
        .withArgs(user1.address, user2.address, transferAmount);

      expect(await wrappedToken.balanceOf(user2.address)).to.equal(
        transferAmount
      );
    });

    it("Should approve and transferFrom correctly", async function () {
      const approveAmount = ethers.parseEther("200");
      const transferAmount = ethers.parseEther("100");

      await wrappedToken.connect(user1).approve(user2.address, approveAmount);
      expect(
        await wrappedToken.allowance(user1.address, user2.address)
      ).to.equal(approveAmount);

      await wrappedToken
        .connect(user2)
        .transferFrom(user1.address, user2.address, transferAmount);
      expect(await wrappedToken.balanceOf(user2.address)).to.equal(
        transferAmount
      );
      expect(
        await wrappedToken.allowance(user1.address, user2.address)
      ).to.equal(approveAmount - transferAmount);
    });
  });
});
