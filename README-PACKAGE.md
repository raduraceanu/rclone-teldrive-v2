# rclone TelDrive Arch package

This repository contains the `teldrive-v2-path-support` rclone fork and an
`rclone-teldrive-git` Arch Linux package definition. The package replaces the
regular `rclone` package and includes the TelDrive v2 API and reverse-proxy
path-prefix support.

The package is built in an Arch Linux container on a Blacksmith runner. To
build the latest package manually, open **Actions → Build Arch package → Run
workflow**. The workflow publishes the `.pkg.tar.zst` file and its SHA256 file
to a prerelease on GitHub.

Install a release directly with:

```sh
sudo pacman -U ./rclone-teldrive-git-*.pkg.tar.zst
```

The `PKGBUILD` and `.SRCINFO` can also be submitted to the AUR. Since this is a
git package, `makepkg -si` will rebuild it from the current `main` branch.
