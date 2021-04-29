# ovsdbviz

## How to run

```
$ GOBIN=`pwd` go install github.com/halfcrazy/ovsdbviz@latest
$ ./ovsdbviz -schema=./examplesvswitch.ovsschema -out=/tmp/ovsdb.dot
$ yum install graphviz
$ dot -Tpng /tmp/ovsdb.dot -o /tmp/ovsdb.png
$ open /tmp/ovsdb.png
```

![OpenVSwitch Schema](https://github.com/halfcrazy/ovsdbviz/blob/master/examples/vswitch.ovsschema.png)
