# Test fixture: a masked credential, declared for one of two reachable hosts.
#
# Both hosts redirect to the same local go-httpbin, so what separates them is
# the credential's own host list and nothing else. The sandbox never receives
# TEST_TOKEN's real value; the proxy substitutes it on the way to httpbin.test
# and leaves it alone on the way to pie.test.
{
  httpbinPort ? "18918",
  pkgs ? import ../../pinned-nixpkgs.nix { },
}:
let
  sandbox = import ../../../default.nix { pkgs = pkgs; };
in
sandbox.mkSandbox {
  pkg = pkgs.bash;
  binName = "bash";
  outName = "sandboxed-bash-masked";
  allowedPackages = [
    pkgs.coreutils
    pkgs.bash
    pkgs.curl
  ];
  allowedDomains = [
    "httpbin.test"
    "pie.test"
  ];
  maskedCredentials = {
    TEST_TOKEN = [ "httpbin.test" ];
  };
  _proxyRedirects = {
    "httpbin.test" = "127.0.0.1:${httpbinPort}";
    "pie.test" = "127.0.0.1:${httpbinPort}";
  };
}
