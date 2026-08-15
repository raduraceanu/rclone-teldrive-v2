# Maintainer: JadenJSJ <jadenjsj@proton.me>

pkgname=rclone-teldrive-git
pkgver=r1.d53e125
pkgrel=1
pkgdesc='Rclone with TelDrive v2 API and proxied path-prefix support'
arch=('x86_64')
url='https://github.com/JSJ-Experiments/rclone-teldrive'
license=('MIT')
provides=('rclone')
conflicts=('rclone')
makedepends=('git' 'go')
source=("rclone-teldrive::git+https://github.com/JSJ-Experiments/rclone-teldrive.git#branch=main")
sha256sums=('SKIP')

pkgver() {
    cd "$srcdir/rclone-teldrive"
    printf 'r%s.%s' "$(git rev-list --count HEAD)" "$(git rev-parse --short HEAD)"
}

build() {
    cd "$srcdir/rclone-teldrive"
    go build -trimpath -ldflags "-s -w -X github.com/rclone/rclone/fs.Version=$pkgver" -o rclone .
}

package() {
    cd "$srcdir/rclone-teldrive"
    install -Dm755 rclone "$pkgdir/usr/bin/rclone"
    install -Dm644 rclone.1 "$pkgdir/usr/share/man/man1/rclone.1"
    install -Dm644 COPYING "$pkgdir/usr/share/licenses/$pkgname/COPYING"
}
