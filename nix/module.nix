{
  config,
  lib,
  pkgs,
  ...
}:

let
  cfg = config.programs.seira;
in
{
  options.programs.seira = {
    enable = lib.mkEnableOption "seira shell script bundler";
    package = lib.mkPackageOption pkgs "seira" { };
  };

  config = lib.mkIf cfg.enable {
    environment.systemPackages = [ cfg.package ];
  };
}
