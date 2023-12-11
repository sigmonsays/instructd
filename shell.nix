{ pkgs ? import <nixpkgs> {} }:

with pkgs;

mkShell {
  buildInputs = [

    git
    gomod2nix

    # Go development
    go_1_20
    gopls
    godef
    gotools

  ];
}
