{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    shared-flake.url = "github:sorafujitani/shared-flake-nix";
  };
  outputs =
    {
      nixpkgs,
      flake-utils,
      shared-flake,
      ...
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShells.default = shared-flake.lib.mkDevShell {
          inherit pkgs;
          name = "wez-kv";
          buildInputs = with pkgs; [
            go_1_26
            goreleaser
            gopls
          ];
        };
      }
    );
}
