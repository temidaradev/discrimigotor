{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    python3
    python3Packages.mutagen
  ];

  shellHook = ''
    echo "Music sorter environment loaded!"
    echo "Run: python main.py"
  '';
}
