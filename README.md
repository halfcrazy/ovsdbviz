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

![ovs vswitch Schema](https://github.com/halfcrazy/ovsdbviz/blob/master/examples/ovs-vswitch.png)

### vtep

![ovs vtep Schema](https://github.com/halfcrazy/ovsdbviz/blob/master/examples/ovs-vtep.png)

### nb

![ovn nb Schema](https://github.com/halfcrazy/ovsdbviz/blob/master/examples/ovn-nb.png)

### sb

![ovn sb Schema](https://github.com/halfcrazy/ovsdbviz/blob/master/examples/ovn-sb.png)

### ic-nb

![ovn ic nb Schema](https://github.com/halfcrazy/ovsdbviz/blob/master/examples/ovn-ic-nb.png)

### ic-sb

![ovn ic sb Schema](https://github.com/halfcrazy/ovsdbviz/blob/master/examples/ovn-ic-sb.png)
