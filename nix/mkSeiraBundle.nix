{ pkgs }:

{
  src,
  pname ? "seira-bundle",
  version ? "0.1.0",
  entrypoint ? null,
  installName ? pname,
  mode ? null,
  shebang ? null,
  type ? null,
  minify ? false,
  treeshake ? false,
  stripComments ? false,
  extraArgs ? [ ],
}:

let
  lib = pkgs.lib;
in
pkgs.stdenvNoCC.mkDerivation {
  inherit pname version src;

  nativeBuildInputs = [ pkgs.seira ];

  buildPhase = ''
    runHook preBuild
    seira build \
      ${lib.optionalString (entrypoint != null) (lib.escapeShellArg entrypoint)} \
      ${lib.optionalString (mode != null) "--mode ${lib.escapeShellArg mode}"} \
      ${lib.optionalString (shebang != null) "--shebang ${lib.escapeShellArg shebang}"} \
      ${lib.optionalString (type != null) "--type ${lib.escapeShellArg type}"} \
      ${lib.optionalString minify "--minify"} \
      ${lib.optionalString treeshake "--treeshake"} \
      ${lib.optionalString stripComments "--strip-comments"} \
      ${lib.concatStringsSep " " (map lib.escapeShellArg extraArgs)} \
      -o bundled.sh
    runHook postBuild
  '';

  installPhase = ''
    runHook preInstall
    mkdir -p $out/bin
    install -m755 bundled.sh $out/bin/${lib.escapeShellArg installName}
    runHook postInstall
  '';
}
