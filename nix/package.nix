{
  lib,
  buildGoModule,
}:

buildGoModule {
  pname = "seira";
  version = "0.0.0-dev";

  src = lib.fileset.toSource {
    root = ../.;
    fileset = lib.fileset.intersection (lib.fileset.gitTracked ../.) (
      lib.fileset.fileFilter (f: !(lib.hasSuffix ".nix" f.name)) ../.
    );
  };

  vendorHash = "sha256-BqB9mGoHcb89oJXRl07EWH90p5Pw8QyLz/uxVCr9RTY=";

  # TODO: テストが安定したら doCheck = true に戻す
  doCheck = false;

  env.CGO_ENABLED = "0";

  meta = {
    description = "Shell script framework and bundler";
    homepage = "https://github.com/Hayao0819/seira";
    mainProgram = "seira";
  };
}
