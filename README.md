# golin

This command swiching the symbolic link of GOROOT

# install

https://github.com/shizuokago/golin/releases

Download the golin that suits your platform.

# use

If "go" command does not exit yet,"install"

    $ golin install {path}

this will install the latest Go on the "path".

a symbolic link calld "current" is created in "path" with the latest version.

    e.g) $ golin install /usr/local/go

```
/usr/local/go
       |-1.15.6
       |-current -> 1.15.6
```

you set the environment variable "GOROOT" to "/usr/local/go/current"

# examples

GOROOT = /usr/local/go/current -> 1.15.6

```
/usr/local/go/current -
       |-1.15.6
       |-current -> 1.15.6
```

    $ golin 1.16

```
/usr/local/go/current -
      |-1.15.6
      |-1.16
      |-current -> 1.16
```

If it does not exist, download it and switch it.

Beta and Candidate Release are also available for install.

e.g.) golin 1.16beta1
      golin 1.16rc1

# when using for the first time

if go already exists...
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

