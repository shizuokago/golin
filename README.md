# golin

This command swiching the symbolic link of GOROOT

```bash
$ go get github.com/shizuokago/golin/cmd/golin
```

created $GOPATH/bin/golin

# examples

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

# when using for the first time

You should create and manage a management directory.


For example ...

If your GOROOT is "/usr/local/go".
This command creates "/usr/local/{version}" and "/usr/local/current".
This will "Hachamecha" "/usr/local"

I think that it is better to set it as "/usr/local/go/{nowversion}".

# super user

It can only be executed by superuser.(symblik link create)

## windows

Please execute command prompt as Administrators.

## other(linux or mac)

```
sudo golin {version}
```

Because there is a possibility that environment variables are not inherited by sudo,please add the following the /etc/sudoers

```
Deafaults env_keep += "GOROOT"
```

