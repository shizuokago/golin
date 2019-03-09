# golin

This command toggles the symbolic ink of  GOROOT

# example

```
target -
       |-1.8.1
       |-1.9
       |-1.11rc1
       |-1.11.4
       |-1.12
       |-current -> 1.11.4
```

```
golin 1.12
```

```
target -
       |-1.11.4
       |-1.12
       |-current -> 1.12
```

# super user

It can only be executed by superuser.

```
sudo golin {version}
```

Because there is a possibility that environment variables are not inherited by sudo,please add the following the /etc/sudoers

```
Deafaults env_keep += "GOROOT"
```

