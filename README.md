# golin

This command swiching the symbolic link of GOROOT

# install

https://github.com/shizuokago/golin/releases

Download the golin that suits your platform.

実行ファイルをPATHに通してください


## already Go Runtime installed.

    go get github.com/shizuokago/golin/v2/_cmd/golin@latest

GOBIN(GOPATH/bin)に配布されます。GOBINにPATHが通っていることを確認してください。

# use

If "go" command does not exit yet,"install"

    $ golin install {path} {version}

this will install the latest Go on the "path".

a symbolic link calld "current" is created in "path" with the latest version.

    e.g) $ golin install /usr/local/go

```
/usr/local/go
       |-1.16.5
       |-current -> 1.16.5
```

you set the environment variable "GOROOT" to "/usr/local/go/current"

# examples

GOROOT = /usr/local/go/current -> 1.16.5

```
/usr/local/go/current -
       |-1.16.5
       |-current -> 1.16.5
```

    $ golin 1.17beta1

```
/usr/local/go/current -
      |-1.16.5
      |-1.17
      |-current -> 1,17
```

If it does not exist, download it and switch it.

Beta and Candidate Release are also available for install.

e.g.) golin 1.17beta1
      golin 1.17rc1

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


# Problem installing v2 with "Modules"

It is known to add v2 to the package "go.mod", but when installing the command, it fits the phenomenon that the package cannot be found.

This is a symptom that only happens if the command is installed in a different location, and I think few packages are currently facing this issue.

Only found

  https://github.com/golang/appengine/tree/master/v2

I tried to refer to the v2 directory of.

