const { expect } = require("chai");
const { ethers } = require("hardhat");

describe("PolygonBridge", function () {
  let polygonBridge;
  let mockERC20;
  let wrappedToken;
  let admin, relayer1, relayer2, relayer3, user1, user2;
  let relayers;

  const ETHEREUM_CHAIN_ID = 1;
  let POLYGON_CHAIN_ID;

  // Helper function for creating signatures
  async function createSignatures(
    transferId,
    originalToken,
    originalChainId,
    amount,
    recipient,
    signers
  ) {
    const messageHash = ethers.solidityPackedKeccak256(
      ["bytes32", "address", "uint256", "uint256", "address", "uint256"],
      [
        transferId,
        originalToken,
        originalChainId,
        amount,
        recipient,
        POLYGON_CHAIN_ID,
      ]
    );

    const signatures = [];
    for (const signer of signers) {
      const signature = await signer.signMessage(ethers.getBytes(messageHash));
      signatures.push(signature);
    }
    return signatures;
  }

  beforeEach(async function () {
    [admin, relayer1, relayer2, relayer3, user1, user2] =
      await ethers.getSigners();
    relayers = [relayer1.address, relayer2.address, relayer3.address];

    // Get the current chain ID
    const network = await ethers.provider.getNetwork();
    POLYGON_CHAIN_ID = Number(network.chainId);

    // Deploy MockERC20 for testing
    const MockERC20 = await ethers.getContractFactory("MockERC20");
    mockERC20 = await MockERC20.deploy("Test Token", "TEST", 18);
    await mockERC20.waitForDeployment();

    // Deploy PolygonBridge
    const PolygonBridge = await ethers.getContractFactory("PolygonBridge");
    polygonBridge = await PolygonBridge.deploy(
      admin.address,
      relayers,
      2 // require 2 signatures
    );
    await polygonBridge.waitForDeployment();
  });

  describe("Deployment", function () {
    it("Should set the correct admin", async function () {
      expect(
        await polygonBridge.hasRole(
          await polygonBridge.ADMIN_ROLE(),
          admin.address
        )
      ).to.be.true;
    });

    it("Should set the correct relayers", async function () {
      const contractRelayers = await polygonBridge.getRelayers();
      expect(contractRelayers).to.deep.equal(relayers);
    });

    it("Should set the correct required signatures", async function () {
      expect(await polygonBridge.requiredSignatures()).to.equal(2);
    });

    it("Should revert with zero admin address", async function () {
      const PolygonBridge = await ethers.getContractFactory("PolygonBridge");
      await expect(
        PolygonBridge.deploy(ethers.ZeroAddress, relayers, 2)
      ).to.be.revertedWithCustomError(polygonBridge, "ZeroAddress");
    });

    it("Should revert with insufficient required signatures", async function () {
      const PolygonBridge = await ethers.getContractFactory("PolygonBridge");
      await expect(
        PolygonBridge.deploy(admin.address, relayers, 0)
      ).to.be.revertedWithCustomError(polygonBridge, "InsufficientSignatures");
    });
  });

  describe("Wrapped Token Deployment", function () {
    it("Should deploy wrapped token successfully", async function () {
      const tx = await polygonBridge.deployWrappedToken(
        mockERC20.target,
        ETHEREUM_CHAIN_ID,
        "Wrapped Test Token",
        "wTEST",
        18
      );

      const receipt = await tx.wait();
      const event = receipt.logs.find((log) => {
        try {
          return (
            polygonBridge.interface.parseLog(log).name ===
            "WrappedTokenDeployed"
          );
        } catch {
          return false;
        }
      });

      expect(event).to.not.be.undefined;
      const parsedEvent = polygonBridge.interface.parseLog(event);

      const wrappedTokenAddress = parsedEvent.args.wrappedToken;
      expect(wrappedTokenAddress).to.not.equal(ethers.ZeroAddress);

      // Verify token info
      const tokenInfo = await polygonBridge.getTokenInfo(
        mockERC20.target,
        ETHEREUM_CHAIN_ID
      );
      expect(tokenInfo.wrappedToken).to.equal(wrappedTokenAddress);
      expect(tokenInfo.originalToken).to.equal(mockERC20.target);
      expect(tokenInfo.originalChainId).to.equal(ETHEREUM_CHAIN_ID);
      expect(tokenInfo.isSupported).to.be.true;
    });

    it("Should revert when deploying token with zero address", async function () {
      await expect(
        polygonBridge.deployWrappedToken(
          ethers.ZeroAddress,
          ETHEREUM_CHAIN_ID,
          "Wrapped Test Token",
          "wTEST",
          18
        )
      ).to.be.revertedWithCustomError(polygonBridge, "ZeroAddress");
    });

    it("Should revert when deploying token for same chain", async function () {
      await expect(
        polygonBridge.deployWrappedToken(
          mockERC20.target,
          POLYGON_CHAIN_ID,
          "Wrapped Test Token",
          "wTEST",
          18
        )
      )
        .to.be.revertedWithCustomError(polygonBridge, "InvalidDestinationChain")
        .withArgs(POLYGON_CHAIN_ID);
    });

    it("Should revert when token already supported", async function () {
      await polygonBridge.deployWrappedToken(
        mockERC20.target,
        ETHEREUM_CHAIN_ID,
        "Wrapped Test Token",
        "wTEST",
        18
      );

      await expect(
        polygonBridge.deployWrappedToken(
          mockERC20.target,
          ETHEREUM_CHAIN_ID,
          "Another Wrapped Token",
          "awTEST",
          18
        )
      ).to.be.revertedWithCustomError(polygonBridge, "TokenAlreadySupported");
    });

    it("Should only allow admin to deploy wrapped tokens", async function () {
      await expect(
        polygonBridge
          .connect(user1)
          .deployWrappedToken(
            mockERC20.target,
            ETHEREUM_CHAIN_ID,
            "Wrapped Test Token",
            "wTEST",
            18
          )
      ).to.be.reverted;
    });
  });

  describe("Token Minting", function () {
    let transferId;
    let wrappedTokenAddress;

    beforeEach(async function () {
      // Deploy wrapped token
      const tx = await polygonBridge.deployWrappedToken(
        mockERC20.target,
        ETHEREUM_CHAIN_ID,
        "Wrapped Test Token",
        "wTEST",
        18
      );

      const receipt = await tx.wait();
      const event = receipt.logs.find((log) => {
        try {
          return (
            polygonBridge.interface.parseLog(log).name ===
            "WrappedTokenDeployed"
          );
        } catch {
          return false;
        }
      });

      wrappedTokenAddress =
        polygonBridge.interface.parseLog(event).args.wrappedToken;
      transferId = ethers.keccak256(ethers.toUtf8Bytes("test-transfer-1"));
    });

    it("Should mint tokens with valid signatures", async function () {
      const amount = ethers.parseEther("100");
      const signatures = await createSignatures(
        transferId,
        mockERC20.target,
        ETHEREUM_CHAIN_ID,
        amount,
        user1.address,
        [relayer1, relayer2]
      );

      await expect(
        polygonBridge.mintTokens(
          transferId,
          mockERC20.target,
          ETHEREUM_CHAIN_ID,
          amount,
          user1.address,
          signatures
        )
      ).to.emit(polygonBridge, "TokensMinted");

      // Check wrapped token balance
      const WrappedToken = await ethers.getContractFactory("WrappedToken");
      const wrappedTokenContract = WrappedToken.attach(wrappedTokenAddress);
      expect(await wrappedTokenContract.balanceOf(user1.address)).to.equal(
        amount
      );

      // Check total minted
      expect(
        await polygonBridge.getTotalMinted(mockERC20.target, ETHEREUM_CHAIN_ID)
      ).to.equal(amount);
    });

    it("Should revert with insufficient signatures", async function () {
      const amount = ethers.parseEther("100");
      const signatures = await createSignatures(
        transferId,
        mockERC20.target,
        ETHEREUM_CHAIN_ID,
        amount,
        user1.address,
        [relayer1] // Only 1 signature, need 2
      );

      await expect(
        polygonBridge.mintTokens(
          transferId,
          mockERC20.target,
          ETHEREUM_CHAIN_ID,
          amount,
          user1.address,
          signatures
        )
      ).to.be.revertedWithCustomError(polygonBridge, "InsufficientSignatures");
    });

    it("Should revert with invalid signatures", async function () {
      const amount = ethers.parseEther("100");
      const signatures = await createSignatures(
        transferId,
        mockERC20.target,
        ETHEREUM_CHAIN_ID,
        amount,
        user1.address,
        [user1, user2] // Non-relayer signatures
      );

      await expect(
        polygonBridge.mintTokens(
          transferId,
          mockERC20.target,
          ETHEREUM_CHAIN_ID,
          amount,
          user1.address,
          signatures
        )
      ).to.be.revertedWithCustomError(polygonBridge, "InsufficientSignatures");
    });

    it("Should revert when transfer already processed", async function () {
      const amount = ethers.parseEther("100");
      const signatures = await createSignatures(
        transferId,
        mockERC20.target,
        ETHEREUM_CHAIN_ID,
        amount,
        user1.address,
        [relayer1, relayer2]
      );

      // First mint
      await polygonBridge.mintTokens(
        transferId,
        mockERC20.target,
        ETHEREUM_CHAIN_ID,
        amount,
        user1.address,
        signatures
      );

      // Second mint should fail
      await expect(
        polygonBridge.mintTokens(
          transferId,
          mockERC20.target,
          ETHEREUM_CHAIN_ID,
          amount,
          user1.address,
          signatures
        )
      ).to.be.revertedWithCustomError(
        polygonBridge,
        "TransferAlreadyProcessed"
      );
    });

    it("Should revert with unsupported token", async function () {
      const amount = ethers.parseEther("100");
      const unsupportedToken = await (
        await ethers.getContractFactory("MockERC20")
      ).deploy("Unsupported", "UNS", 18);

      const signatures = await createSignatures(
        transferId,
        unsupportedToken.target,
        ETHEREUM_CHAIN_ID,
        amount,
        user1.address,
        [relayer1, relayer2]
      );

      await expect(
        polygonBridge.mintTokens(
          transferId,
          unsupportedToken.target,
          ETHEREUM_CHAIN_ID,
          amount,
          user1.address,
          signatures
        )
      ).to.be.revertedWithCustomError(polygonBridge, "TokenNotSupported");
    });

    it("Should revert with zero amount", async function () {
      const signatures = await createSignatures(
        transferId,
        mockERC20.target,
        ETHEREUM_CHAIN_ID,
        0,
        user1.address,
        [relayer1, relayer2]
      );

      await expect(
        polygonBridge.mintTokens(
          transferId,
          mockERC20.target,
          ETHEREUM_CHAIN_ID,
          0,
          user1.address,
          signatures
        )
      ).to.be.revertedWithCustomError(polygonBridge, "InsufficientAmount");
    });

    it("Should revert with zero recipient address", async function () {
      const amount = ethers.parseEther("100");
      const signatures = await createSignatures(
        transferId,
        mockERC20.target,
        ETHEREUM_CHAIN_ID,
        amount,
        ethers.ZeroAddress,
        [relayer1, relayer2]
      );

      await expect(
        polygonBridge.mintTokens(
          transferId,
          mockERC20.target,
          ETHEREUM_CHAIN_ID,
          amount,
          ethers.ZeroAddress,
          signatures
        )
      ).to.be.revertedWithCustomError(polygonBridge, "ZeroAddress");
    });
  });

  describe("Token Burning", function () {
    let wrappedTokenAddress;
    let wrappedTokenContract;

    beforeEach(async function () {
      // Deploy wrapped token
      const tx = await polygonBridge.deployWrappedToken(
        mockERC20.target,
        ETHEREUM_CHAIN_ID,
        "Wrapped Test Token",
        "wTEST",
        18
      );

      const receipt = await tx.wait();
      const event = receipt.logs.find((log) => {
        try {
          return (
            polygonBridge.interface.parseLog(log).name ===
            "WrappedTokenDeployed"
          );
        } catch {
          return false;
        }
      });

      wrappedTokenAddress =
        polygonBridge.interface.parseLog(event).args.wrappedToken;

      const WrappedToken = await ethers.getContractFactory("WrappedToken");
      wrappedTokenContract = WrappedToken.attach(wrappedTokenAddress);

      // Mint some tokens to user1 for burning tests
      const transferId = ethers.keccak256(ethers.toUtf8Bytes("mint-for-burn"));
      const amount = ethers.parseEther("1000");

      const messageHash = ethers.solidityPackedKeccak256(
        ["bytes32", "address", "uint256", "uint256", "address", "uint256"],
        [
          transferId,
          mockERC20.target,
          ETHEREUM_CHAIN_ID,
          amount,
          user1.address,
          POLYGON_CHAIN_ID,
        ]
      );

      const signatures = [];
      for (const signer of [relayer1, relayer2]) {
        const signature = await signer.signMessage(
          ethers.getBytes(messageHash)
        );
        signatures.push(signature);
      }

      await polygonBridge.mintTokens(
        transferId,
        mockERC20.target,
        ETHEREUM_CHAIN_ID,
        amount,
        user1.address,
        signatures
      );
    });

    it("Should burn tokens successfully", async function () {
      const burnAmount = ethers.parseEther("100");
      const initialBalance = await wrappedTokenContract.balanceOf(
        user1.address
      );

      const tx = await polygonBridge
        .connect(user1)
        .burnTokens(
          wrappedTokenAddress,
          burnAmount,
          ETHEREUM_CHAIN_ID,
          user2.address
        );

      const receipt = await tx.wait();
      const event = receipt.logs.find((log) => {
        try {
          return polygonBridge.interface.parseLog(log).name === "TokensBurned";
        } catch {
          return false;
        }
      });

      expect(event).to.not.be.undefined;
      const parsedEvent = polygonBridge.interface.parseLog(event);

      expect(parsedEvent.args.user).to.equal(user1.address);
      expect(parsedEvent.args.wrappedToken).to.equal(wrappedTokenAddress);
      expect(parsedEvent.args.originalToken).to.equal(mockERC20.target);
      expect(parsedEvent.args.amount).to.equal(burnAmount);
      expect(parsedEvent.args.destinationChain).to.equal(ETHEREUM_CHAIN_ID);
      expect(parsedEvent.args.recipient).to.equal(user2.address);

      // Check balance decreased
      expect(await wrappedTokenContract.balanceOf(user1.address)).to.equal(
        initialBalance - burnAmount
      );

      // Check total minted decreased
      expect(
        await polygonBridge.getTotalMinted(mockERC20.target, ETHEREUM_CHAIN_ID)
      ).to.equal(initialBalance - burnAmount);
    });

    it("Should revert with zero amount", async function () {
      await expect(
        polygonBridge
          .connect(user1)
          .burnTokens(wrappedTokenAddress, 0, ETHEREUM_CHAIN_ID, user2.address)
      ).to.be.revertedWithCustomError(polygonBridge, "InsufficientAmount");
    });

    it("Should revert with zero recipient address", async function () {
      await expect(
        polygonBridge
          .connect(user1)
          .burnTokens(
            wrappedTokenAddress,
            ethers.parseEther("100"),
            ETHEREUM_CHAIN_ID,
            ethers.ZeroAddress
          )
      ).to.be.revertedWithCustomError(polygonBridge, "ZeroAddress");
    });

    it("Should revert with same chain as destination", async function () {
      await expect(
        polygonBridge
          .connect(user1)
          .burnTokens(
            wrappedTokenAddress,
            ethers.parseEther("100"),
            POLYGON_CHAIN_ID,
            user2.address
          )
      ).to.be.revertedWithCustomError(polygonBridge, "InvalidDestinationChain");
    });

    it("Should revert with unsupported wrapped token", async function () {
      const unsupportedToken = await (
        await ethers.getContractFactory("MockERC20")
      ).deploy("Unsupported", "UNS", 18);

      await expect(
        polygonBridge
          .connect(user1)
          .burnTokens(
            unsupportedToken.target,
            ethers.parseEther("100"),
            ETHEREUM_CHAIN_ID,
            user2.address
          )
      ).to.be.revertedWithCustomError(polygonBridge, "WrappedTokenNotFound");
    });

    it("Should revert when user has insufficient balance", async function () {
      const userBalance = await wrappedTokenContract.balanceOf(user1.address);
      const burnAmount = userBalance + ethers.parseEther("1");

      await expect(
        polygonBridge
          .connect(user1)
          .burnTokens(
            wrappedTokenAddress,
            burnAmount,
            ETHEREUM_CHAIN_ID,
            user2.address
          )
      ).to.be.revertedWithPanic(0x11); // Arithmetic overflow
    });
  });

  describe("Token Registry Management", function () {
    it("Should add supported token", async function () {
      // Deploy a wrapped token externally
      const WrappedToken = await ethers.getContractFactory("WrappedToken");
      const externalWrappedToken = await WrappedToken.deploy(
        "External Wrapped Token",
        "EWT",
        18,
        mockERC20.target,
        ETHEREUM_CHAIN_ID,
        polygonBridge.target
      );

      await polygonBridge.addSupportedToken(
        mockERC20.target,
        ETHEREUM_CHAIN_ID,
        externalWrappedToken.target
      );

      const tokenInfo = await polygonBridge.getTokenInfo(
        mockERC20.target,
        ETHEREUM_CHAIN_ID
      );
      expect(tokenInfo.isSupported).to.be.true;
      expect(tokenInfo.wrappedToken).to.equal(externalWrappedToken.target);
    });

    it("Should remove supported token", async function () {
      // First deploy and add a token
      await polygonBridge.deployWrappedToken(
        mockERC20.target,
        ETHEREUM_CHAIN_ID,
        "Wrapped Test Token",
        "wTEST",
        18
      );

      // Remove support
      await polygonBridge.removeSupportedToken(
        mockERC20.target,
        ETHEREUM_CHAIN_ID
      );

      const tokenInfo = await polygonBridge.getTokenInfo(
        mockERC20.target,
        ETHEREUM_CHAIN_ID
      );
      expect(tokenInfo.isSupported).to.be.false;
    });

    it("Should only allow admin to manage token registry", async function () {
      await expect(
        polygonBridge
          .connect(user1)
          .removeSupportedToken(mockERC20.target, ETHEREUM_CHAIN_ID)
      ).to.be.reverted;
    });
  });

  describe("Relayer Management", function () {
    it("Should add new relayer", async function () {
      await polygonBridge.addRelayer(user1.address);
      expect(
        await polygonBridge.hasRole(
          await polygonBridge.RELAYER_ROLE(),
          user1.address
        )
      ).to.be.true;

      const relayersList = await polygonBridge.getRelayers();
      expect(relayersList).to.include(user1.address);
    });

    it("Should remove relayer", async function () {
      await polygonBridge.removeRelayer(relayer1.address);
      expect(
        await polygonBridge.hasRole(
          await polygonBridge.RELAYER_ROLE(),
          relayer1.address
        )
      ).to.be.false;

      const relayersList = await polygonBridge.getRelayers();
      expect(relayersList).to.not.include(relayer1.address);
    });

    it("Should update required signatures", async function () {
      await polygonBridge.updateRequiredSignatures(1);
      expect(await polygonBridge.requiredSignatures()).to.equal(1);
    });

    it("Should revert when setting required signatures to zero", async function () {
      await expect(
        polygonBridge.updateRequiredSignatures(0)
      ).to.be.revertedWithCustomError(polygonBridge, "InsufficientSignatures");
    });

    it("Should only allow admin to manage relayers", async function () {
      await expect(polygonBridge.connect(user1).addRelayer(user2.address)).to.be
        .reverted;
    });
  });

  describe("Pause/Unpause", function () {
    it("Should pause and unpause contract", async function () {
      await polygonBridge.pause();
      expect(await polygonBridge.paused()).to.be.true;

      await polygonBridge.unpause();
      expect(await polygonBridge.paused()).to.be.false;
    });

    it("Should prevent operations when paused", async function () {
      await polygonBridge.pause();

      // Deploy wrapped token first
      await polygonBridge.unpause();
      await polygonBridge.deployWrappedToken(
        mockERC20.target,
        ETHEREUM_CHAIN_ID,
        "Wrapped Test Token",
        "wTEST",
        18
      );
      await polygonBridge.pause();

      // Test that minting is paused
      const transferId = ethers.keccak256(
        ethers.toUtf8Bytes("test-transfer-paused")
      );
      const amount = ethers.parseEther("100");
      const signatures = await createSignatures(
        transferId,
        mockERC20.target,
        ETHEREUM_CHAIN_ID,
        amount,
        user1.address,
        [relayer1, relayer2]
      );

      await expect(
        polygonBridge.mintTokens(
          transferId,
          mockERC20.target,
          ETHEREUM_CHAIN_ID,
          amount,
          user1.address,
          signatures
        )
      ).to.be.revertedWithCustomError(polygonBridge, "EnforcedPause");
    });

    it("Should only allow admin to pause/unpause", async function () {
      await expect(polygonBridge.connect(user1).pause()).to.be.reverted;
    });
  });

  describe("View Functions", function () {
    let wrappedTokenAddress;

    beforeEach(async function () {
      const tx = await polygonBridge.deployWrappedToken(
        mockERC20.target,
        ETHEREUM_CHAIN_ID,
        "Wrapped Test Token",
        "wTEST",
        18
      );

      const receipt = await tx.wait();
      const event = receipt.logs.find((log) => {
        try {
          return (
            polygonBridge.interface.parseLog(log).name ===
            "WrappedTokenDeployed"
          );
        } catch {
          return false;
        }
      });

      wrappedTokenAddress =
        polygonBridge.interface.parseLog(event).args.wrappedToken;
    });

    it("Should get wrapped token address", async function () {
      const wrappedToken = await polygonBridge.getWrappedToken(
        mockERC20.target,
        ETHEREUM_CHAIN_ID
      );
      expect(wrappedToken).to.equal(wrappedTokenAddress);
    });

    it("Should get original token info", async function () {
      const [originalToken, originalChainId] =
        await polygonBridge.getOriginalToken(wrappedTokenAddress);
      expect(originalToken).to.equal(mockERC20.target);
      expect(originalChainId).to.equal(ETHEREUM_CHAIN_ID);
    });

    it("Should check if token is supported", async function () {
      expect(
        await polygonBridge.isTokenSupported(
          mockERC20.target,
          ETHEREUM_CHAIN_ID
        )
      ).to.be.true;
      expect(await polygonBridge.isTokenSupported(mockERC20.target, 999)).to.be
        .false;
    });

    it("Should get relayers list", async function () {
      const relayersList = await polygonBridge.getRelayers();
      expect(relayersList).to.deep.equal(relayers);
    });
  });
});
