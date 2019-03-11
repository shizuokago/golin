# golin

This command toggles the symbolic link of GOROOT

```bash
$ go get github.com/shizuokago/golin/cmd/golin
```

created $GOPATH/bin/golin

# example

GOROOT = /usr/local/go/current -> 1.11.4

```
/usr/local/go/current -
       |-1.11rc1
       |-1.11.4
       |-current -> 1.11.4
```

```bash
$ golin 1.12
```

```
/usr/local/go/current -
       |-1.11rc1
       |-1.11.4
       |-1.12
       |-current -> 1.12
```

If it does not exist, download it and switch it.


# super user

It can only be executed by superuser.

## windows

Please execute command prompt as Administrators.

## other

```
sudo golin {version}
```

Because there is a possibility that environment variables are not inherited by sudo,please add the following the /etc/sudoers

```
Deafaults env_keep += "GOROOT"
```

