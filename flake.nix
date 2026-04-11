{
  description = "Seira - shell script bundler";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

  outputs =
    { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
      pkgsFor = system: import nixpkgs { inherit system; overlays = [ self.overlays.default ]; };
    in
    {
      lib.mkSeiraBundle = import ./nix/mkSeiraBundle.nix;

      overlays.default = final: prev: {
        seira = final.callPackage ./nix/package.nix { };
      };

      packages = forAllSystems (
        system:
        let
          pkgs = pkgsFor system;
        in
        {
          seira = pkgs.seira;
          default = pkgs.seira;
        }
      );

      devShells = forAllSystems (
        system:
        let
          pkgs = pkgsFor system;
        in
        {
          default = pkgs.mkShell {
            inputsFrom = [ pkgs.seira ];
            packages = with pkgs; [
              gopls
              gotools
              go-tools
            ];
          };
        }
      );

      nixosModules.default = import ./nix/module.nix;
    };
}
