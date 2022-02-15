{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  packages = with pkgs; [
    # Go
    go_1_17
    goimports
    # Rust
    cargo
    rustfmt
    rustc
    # Python
    python3
    python3Packages.black
    # SQLite
    sqlite
    # Formatters
    nodePackages.prettier
    shfmt
  ];
}
