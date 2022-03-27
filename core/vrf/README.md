Build libsodium
```go
aclocal
autoconf
automake --add-missing
./configure
make


copy libsodium.a to ${SRCDIR}/libs/darwin/amd64/lib/libsodium.a
```