# ovsdbviz

**[Interactive demo](https://halfcrazy.github.io/ovsdbviz/)** — all schemas from the latest OVS/OVN trees, rendered in your browser (requires GitHub Pages enabled: Settings → Pages → Deploy from branch → `master` → `/docs`).

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

`--format html` produces a single self-contained HTML file (Cytoscape.js + ELK layered layout via CDN, no build step):

- Multiple schemas in one page: repeat `--schema` or pass files as positional args; switch schemas from the dropdown
- Click a table: keeps only its direct & transitive relations and re-layouts the subgraph (fewer crossings); click blank space to restore
- Edges leave from the exact field row, like Graphviz record ports; red = key reference, blue = value reference, dotted = weak reference (RFC 7047)
- "Show all columns" renders every column as a record row; by default only referencing columns are shown
- "Load JSON…" adds schemas generated with `--format json` at runtime

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
