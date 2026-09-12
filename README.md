# ovsdbviz

## How to run

```
$ GOBIN=`pwd` go install github.com/halfcrazy/ovsdbviz@latest
$ ./ovsdbviz --schema ./examplesvswitch.ovsschema --out ./ovsdb1.dot
$ ./ovsdbviz --db OVN_Northbound --address 192.168.1.1:6641 --out ./ovsdb2.dot
$ yum install graphviz
$ dot -Tpng ./ovsdb.dot -o ./ovsdb.png
$ open ./ovsdb.png
```

## Interactive HTML output

`--format html` 生成单个自包含 HTML（Cytoscape.js + ELK 布局，经 CDN 加载，无需构建）：

- 支持多个 schema（`--schema` 可重复，或直接跟多个文件参数），页面顶部下拉切换
- 点击表名节点：高亮其所有直接/间接（传递闭包）关联的表，弱化其余；勾选"隐藏无关"则直接隐藏
- 点击空白处重置高亮；点击节点后右侧栏展示该表全部列及类型
- 边的颜色：红色 = key 引用，蓝色 = value 引用，点线 = weak reference（RFC 7047）
- 页面上的"加载 JSON…"可追加加载 `--format json` 生成的 schema 文件

```
$ ./ovsdbviz --format html --out ./ovsdb.html examples/*.ovsschema
$ open ./ovsdb.html
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
