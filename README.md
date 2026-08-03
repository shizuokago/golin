# golin

This command switches the symbolic link of GOROOT

# install

https://github.com/shizuokago/golin/releases

Download the golin that suits your platform.

実行ファイルをPATHに通してください


## already Go Runtime installed.

    go install github.com/shizuokago/golin/v2/_cmd/golin@latest

GOBIN(GOPATH/bin)に配布されます。GOBINにPATHが通っていることを確認してください。

# use

If "go" command does not exist yet,"install"

    $ golin install {path} {version}

this will install Go on the "path".

{version}を省略した場合、最新のバージョンがインストールされます。

a symbolic link called "current" is created in "path" with the specified version.

**{path}は事前に作成しておいてください。golinは{path}自体の作成を行いません。**

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
/usr/local/go
       |-1.16.5
       |-current -> 1.16.5
```

    $ golin 1.17beta1

```
/usr/local/go
      |-1.16.5
      |-1.17beta1
      |-current -> 1.17beta1
```

If it does not exist, download it and switch it.

Beta and Candidate Release are also available for install.

e.g.) golin 1.17beta1
      golin 1.17rc1

# other commands

## list

    $ golin list

インストール可能なバージョンをマイナーバージョンごとにまとめて表示します。
既にインストール済みのバージョンはインデントして表示されます。
一覧は https://go.dev/dl/ から取得する為、ネットワークが必要です。

    $ golin -all list

こちらはまとめずに全バージョンを1行ずつ表示します。

オプションはサブコマンドより前に指定してください。
`golin list -all`ではオプションとして認識されません。

## dev / tip

    $ golin dev

現在開発中の最新バージョン(gotip)をソースからビルドして切り替えます。
少し時間がかかります。devは既にビルド済みであっても常にビルドし直します。

    $ golin tip

devでビルドしたものが存在する場合、ビルドを行わず切り替えるのみで終了します。
存在しない場合はdevと同様にビルドを行います。

## version

    $ golin version

golin自体のバージョンを表示します。

# option

## -d

作成するシンボリックリンクの名前を"current"から変更します。

    e.g.) $ golin -d root 1.16

```
/usr/local/go
      |-1.16
      |-root -> 1.16
```

この場合、環境変数GOROOTには"/usr/local/go/root"を設定してください。

# super user

It can only be executed by superuser.(symbolic link create)

## windows

Please execute command prompt as Administrators.

## other(linux or mac)

```
sudo golin {version}
```

Because there is a possibility that environment variables are not inherited by sudo,please add the following the /etc/sudoers

```
Defaults env_keep += "GOROOT"
```


# Problem installing v2 with "Modules"

It is known to add v2 to the package "go.mod", but when installing the command, it fits the phenomenon that the package cannot be found.

This is a symptom that only happens if the command is installed in a different location, and I think few packages are currently facing this issue.

Only found

  https://github.com/golang/appengine/tree/master/v2

I tried to refer to the v2 directory of.


