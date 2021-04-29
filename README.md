# ovsdbviz

## How to run

```
$ GOBIN=`pwd` go install github.com/halfcrazy/ovsdbviz@latest
$ ./ovsdbviz -schema ./examplesvswitch.ovsschema -out ./ovsdb.dot
$ yum install graphviz
$ dot -Tpng ./ovsdb.dot -o ./ovsdb.png
$ open ./ovsdb.png
```

![OpenVSwitch Schema](https://github.com/halfcrazy/ovsdbviz/blob/master/examples/vswitch.ovsschema.png)
