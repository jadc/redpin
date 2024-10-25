{
    description = "redpin dev environment";
    inputs.nixpkgs.url = "nixpkgs/nixos-unstable";
    outputs = { nixpkgs }: let
        system = "x86_64-linux";
        pkgs = import nixpkgs { inherit system; };
    in {
        devShells.${system}.default = pkgs.mkShell {
            packages = with pkgs; [ go ];
            VERBOSE = 1;
        };
    };
}