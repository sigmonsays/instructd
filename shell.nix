{ pkgs ? import <nixpkgs> { } }:

with pkgs;

mkShell {
  buildInputs = [
    git
    #gomod2nix
    nixpkgs-fmt

    # Go development
    go
    gopls
    godef
    gotools

    mpg123
    openssl

  ];
}
