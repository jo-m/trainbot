{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.11";
    unstable.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  };

  outputs = {
    self,
    nixpkgs,
    unstable,
  } @ inputs: let
    lib = nixpkgs.lib;
    system = "x86_64-linux";
    pkgs = import inputs.nixpkgs {
      system = system;
    };
    unstablePkgs = import inputs.unstable {
      system = system;
    };
  in {
    formatter.${system} = nixpkgs.legacyPackages.${system}.alejandra;

    devShells.${system}.default = pkgs.mkShell rec {
      nativeBuildInputs = with pkgs; [];
      packages = with pkgs; [];

      buildInputs = with pkgs; [
        # Build tools
        gcc
        pkg-config
        clang-tools
        gnumake
        curl
        unstablePkgs.go_1_25

        # Cross
        pkgsCross.aarch64-multiplatform.buildPackages.gcc # Provides aarch64-unknown-linux-gnu-gcc

        # Vulkan bare tools and dependencies
        glslang
        vulkan-headers
        vulkan-loader
        vulkan-validation-layers

        # More Vulkan tools
        vulkan-extension-layer
        vulkan-tools
        vulkan-tools-lunarg
        vulkan-volk

        # Frontend
        nodejs_24
      ];

      LD_LIBRARY_PATH = "${lib.makeLibraryPath buildInputs}";
      VK_LAYER_PATH = "${pkgs.vulkan-validation-layers}/share/vulkan/explicit_layer.d";
      VULKAN_SDK = "${pkgs.vulkan-validation-layers}/share/vulkan/explicit_layer.d";
      XDG_DATA_DIRS = builtins.getEnv "XDG_DATA_DIRS";
      XDG_RUNTIME_DIR = "/run/user/1000";
    };
  };
}
