# ovsdbviz

## How to run

```
$ GOBIN=`pwd` go install github.com/halfcrazy/ovsdbviz@latest
$ ./ovsdbviz -schema ./examplesvswitch.ovsschema -out ./ovsdb.dot
$ yum install graphviz
$ dot -Tpng ./ovsdb.dot -o ./ovsdb.png
$ open ./ovsdb.png
```

### vswitch

![ovs vswitch Schema](https://github.com/halfcrazy/ovsdbviz/blob/master/examples/vswitch.png)

### vtep

![ovs vtep Schema](https://github.com/halfcrazy/ovsdbviz/blob/master/examples/vtep.png)

### nb

![ovn nb Schema](https://github.com/halfcrazy/ovsdbviz/blob/master/examples/nb.png)

### sb

![ovn sb Schema](https://github.com/halfcrazy/ovsdbviz/blob/master/examples/sb.png)
