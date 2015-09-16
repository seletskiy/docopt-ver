docopt-ver is small tool, which will parse specified Go-file and replace
hardcoded version string with specified string.

Most usable as pre-commit hook.

# Installation

## For new git-repos

```bash
go get github.com/seletskiy/docopt-ver

mkdir -p ~/.git/templates/hooks
git config --global init.templatedir ~/.git/templates/
cp $GOPATH/src/github.com/seletskiy/docopt-ver/pre-commit ~/.git/templates/hooks
```

## For existing git-repos

Just copy `pre-commit` file to the `.git/hooks` directory.
