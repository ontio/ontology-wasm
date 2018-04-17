# wasm binary encoding

## Data types

### Numbers

#### `uintN`
N(8,16,32) bits无符号整型，N/8 bytes 低位优先

#### `varuintN`
LEB( **Little Endian Base 128**)128 变长整型 ，最长N bits ([0, 2^*N*-1])，代表最长 ceil(*N*/7) bytes 并可能包含 `0x80` 的补位bytes

#### `varintN`

Signed LEB128 变长整型， 最长N bits ( [-2^(*N*-1), +2^(*N*-1)-1])代表最长 ceil(*N*/7) bytes 并可能包含 `0x80` 或`oxFF`的补位bytes



## Instruction Opcodes

当前所有的指令操作码都是使用1个byte编码表示

## Language Types

所有的类型使用标识为一个负数的`varint7`值，作为类型构造的第一个byte

| Opcode                          | Type constructor    |
| ------------------------------- | ------------------- |
| `-0x01` (i.e., the byte `0x7f`) | `i32`               |
| `-0x02` (i.e., the byte `0x7e`) | `i64`               |
| `-0x03` (i.e., the byte `0x7d`) | `f32`               |
| `-0x04` (i.e., the byte `0x7c`) | `f64`               |
| `-0x10` (i.e., the byte `0x70`) | `anyfunc`           |
| `-0x20` (i.e., the byte `0x60`) | `func`              |
| `-0x40` (i.e., the byte `0x40`) | 代表一个空的 `block_type` |



#### `value_type`

value type 使用一个`varint7`表示

- `i32`
- `i64`
- `f32`
- `f64`



#### `block_type`

block signature 使用一个`varint7`表示，这些类型也编码为：

- 使用value_type 表示签名和单一返回值（a signature with a single result）

- `-0x40` (i.e., the byte `0x40`) 代表一个签名和无返回值

  ​

#### `elem_type`

使用一个`varint7`表示在 table 中的元素类型，当前MVP版本只有 anyfunc



#### `func_type`

函数签名描述，类型的构造体在之后有追加的描述（ Its type constructor is followed by an additional description）

| Field        | Type          | Description                              |
| ------------ | ------------- | ---------------------------------------- |
| form         | `varint7`     | the value for the `func` type constructor as defined above |
| param_count  | `varuint32`   | 函数的参数个数                                  |
| param_types  | `value_type*` | 函数的参数类型                                  |
| return_count | `varuint1`    | 函数返回值个数                                  |
| return_type  | `value_type?` | 函数返回值类型（如果返回值个数是1）                       |



### Other Types

#### `global_type`

| Field        | Type         | Description      |
| ------------ | ------------ | ---------------- |
| content_type | `value_type` | 值的类型             |
| mutability   | `varuint1`   | `0` 不可变, `1`  可变 |



#### `table_type`

| Field        | Type               | Description       |
| ------------ | ------------------ | ----------------- |
| element_type | `elem_type`        | 元素类型              |
| limits       | `resizable_limits` | 见resizable_limits |



#### `memory_type`

| Field  | Type               | Description       |
| ------ | ------------------ | ----------------- |
| limits | `resizable_limits` | 见resizable_limits |



#### `external_kind`

使用1个byte 的无符号整型表示定义是否被imported or defined

- `0` indicating a `Function` import or definition
- `1` indicating a `Table` import or definition
- `2` indicating a `Memory` import or definition
- `3` indicating a `Global` import or definition



#### `resizable_limits`

| Field   | Type         | Description                              |
| ------- | ------------ | ---------------------------------------- |
| flags   | `varuint1`   | `1` if the maximum field is present, `0` otherwise |
| initial | `varuint32`  | initial length (in units of table elements or wasm pages) |
| maximum | `varuint32`? | only present if specified by `flags`     |



#### `init_expr`

intializer_expression 的编码，是表达式的正常编码，以`end`操作码作为分隔符。

注意`get_global`在一个初始化表达式只能指不变的imported全局变量，所有使用的`init_expr`只能出现在imports section。



# Module structure

## High-level structure

module 的开头包括两个字段

| Field        | Type     | Description                              |
| ------------ | -------- | ---------------------------------------- |
| magic number | `uint32` | Magic number `0x6d736100` (i.e., ‘\0asm’) |
| version      | `uint32` | Version number, `0x1`                    |

module 的开头之后是一系列的sections，每个section被定义为一个1byte的*section code*  ，代表一个known section或者是custom setion，接下来是section的长度和payload data，Known section包含一些非零的id，custom sections 包含一个 `0` 的id和后面的识别字符串作为payload 的一部分。

| Field        | Type         | Description                              |
| ------------ | ------------ | ---------------------------------------- |
| id           | `varuint7`   | section code                             |
| payload_len  | `varuint32`  | section长度（bytes）                         |
| name_len     | `varuint32`? | length of `name` in bytes, present if `id == 0` |
| name         | `bytes` ?    | section 名称: valid UTF-8 byte sequence, present if `id == 0` |
| payload_data | `bytes`      | section内容, 长度为 `payload_len - sizeof(name) - sizeof(name_len)` |



每个known section 都是可选的并且最多出现一次，Custom sections都有一个相同的`id` (0)，可以命名为非唯一的(所有组成它们名称的字节可能是相同的)。

Custom setions 将用于调试信息、未来的演进或第三方扩展。对于MVP，我们使用一个特定的自定义部分(Name Section)来调试信息。

如果一个WebAssembly实现的任何custom setion 的payload在验证和编译过程中发生了错误，那么这个错误必须不能使module失效（, errors in that payload must not invalidate the module.）

Known sections 如下表所示，必须按顺序出现，custom section 则可以任意出现在每个known section的前，中，后的位置，某些custom section可能有它们自己的顺序和基数要求。例如，在Data section之后，Name section最多只出现一次。违反这些要求最多可能会导致执行忽略该部分，而不会使模块失效。

每个section的内容编码为 `payload_data`.

| Section Name | Code | Description                              |
| ------------ | ---- | ---------------------------------------- |
| Type         | `1`  | 函数的签名声明                                  |
| Import       | `2`  | Import 声明                                |
| Function     | `3`  | 函数声明                                     |
| Table        | `4`  | Indirect function table and other tables |
| Memory       | `5`  | Memory attributes                        |
| Global       | `6`  | 全局变量声明                                   |
| Export       | `7`  | 导出                                       |
| Start        | `8`  | Start function declaration               |
| Element      | `9`  | Elements section                         |
| Code         | `10` | Function bodies (code)                   |
| Data         | `11` | Data segments                            |

最后一个section的结尾必须与模块的最后一个字节一致。最短的有效模块是8个字节(`magic number`, `version`，后面是zero section)。

### Type section

声明本module内使用的所有函数的签名

| Field   | Type         | Description     |
| ------- | ------------ | --------------- |
| count   | `varuint32`  | type entries数量。 |
| entries | `func_type*` | 见之前的func_type描述 |



### Import section

声明本module内使用的所有imports

| Field   | Type            | Description       |
| ------- | --------------- | ----------------- |
| count   | `varuint32`     | import entries 数量 |
| entries | `import_entry*` | 如下import entry    |

#### Import entry

| Field      | Type            | Description                            |
| ---------- | --------------- | -------------------------------------- |
| module_len | `varuint32`     | length of `module_str` in bytes        |
| module_str | `bytes`         | module name: valid UTF-8 byte sequence |
| field_len  | `varuint32`     | length of `field_str` in bytes         |
| field_str  | `bytes`         | field name: valid UTF-8 byte sequence  |
| kind       | `external_kind` | the kind of definition being imported  |

如果kind 是 `Function`

| Field | Type        | Description                          |
| ----- | ----------- | ------------------------------------ |
| type  | `varuint32` | type index of the function signature |

如果kind是`Table`

| Field | Type         | Description                |
| ----- | ------------ | -------------------------- |
| type  | `table_type` | type of the imported table |

如果kind是 `Memory`

| Field | Type          | Description                 |
| ----- | ------------- | --------------------------- |
| type  | `memory_type` | type of the imported memory |

如果kind是`Global`

| Field | Type          | Description                 |
| ----- | ------------- | --------------------------- |
| type  | `global_type` | type of the imported global |

注意，MVP中只有不可变的全局变量可以被引入。

### Function section

声明所有本module中的函数签名（函数的定义在code sections)

| Field | Type         | Description                              |
| ----- | ------------ | ---------------------------------------- |
| count | `varuint32`  | 函数签名的数量                                  |
| types | `varuint32*` | sequence of indices into the type section |

### Table section

| Field   | Type          | Description    |
| ------- | ------------- | -------------- |
| count   | `varuint32`   | 本模块中定义的table数量 |
| entries | `table_type*` | 见 `table_type` |

MVP中，table的数量不多于1.

### Memory section

ID: `memory`

| Field   | Type           | Description     |
| ------- | -------------- | --------------- |
| count   | `varuint32`    | 本模块中定义的memory数量 |
| entries | `memory_type*` | 见`memory_type`  |

注意 initial/maxium字段被指定为 WebAssembly pages( that the initial/maximum fields are specified in units of [WebAssembly pages](http://webassembly.org/docs/semantics/#linear-memory).)

MVP 中，memory的数量不多于1.

### Global section

| Field   | Type               | Description |
| ------- | ------------------ | ----------- |
| count   | `varuint32`        | 全局变量数量      |
| globals | `global_variable*` | 如下          |

#### Global Entry

每一个全局变量`global_variable` ， 使用给定的类型，可变性和初始值来声明一个全局变量

| Field | Type          | Description |
| ----- | ------------- | ----------- |
| type  | `global_type` | 变量类型        |
| init  | `init_expr`   | 初始值         |

MVP中，只有不可变全局变量可以导出。

### Export section

| Field   | Type            | Description      |
| ------- | --------------- | ---------------- |
| count   | `varuint32      | export entries数量 |
| entries | `export_entry*` | 如下               |

#### Export entry

| Field     | Type            | Description                              |
| --------- | --------------- | ---------------------------------------- |
| field_len | `varuint32`     | length of `field_str` in bytes           |
| field_str | `bytes`         | field name: valid UTF-8 byte sequence    |
| kind      | `external_kind` | the kind of definition being exported    |
| index     | `varuint32`     | the index into the corresponding `index space` |

例如，如果`kind`是`Function`， "index"就是一个`function index` ，在MVP中，memory和table导出的 index 只有0是有效的。

### 

### Start section

声明 [start function](http://webassembly.org/docs/modules/#module-start-function).

| Field | Type        | Description                              |
| ----- | ----------- | ---------------------------------------- |
| index | `varuint32` | start [function index](http://webassembly.org/docs/modules/#function-index-space) |

### Element section

| Field   | Type            | Description |
| ------- | --------------- | ----------- |
| count   | `varuint32`     | 元素段数量       |
| entries | `elem_segment*` | 如下          |

a `elem_segment` is:

| Field    | Type         | Description                              |
| -------- | ------------ | ---------------------------------------- |
| index    | `varuint32`  | the [table index](http://webassembly.org/docs/modules/#table-index-space) (0 in the MVP) |
| offset   | `init_expr`  | an `i32` initializer expression that computes the offset at which to place the elements |
| num_elem | `varuint32`  | number of elements to follow             |
| elems    | `varuint32*` | sequence of [function indices](http://webassembly.org/docs/modules/#function-index-space) |

### Code section

ID: `code`

The code section contains a body for every function in the module. The count of function declared in the [function section](http://webassembly.org/docs/binary-encoding/#function-section) and function bodies defined in this section must be the same and the `i`th declaration corresponds to the `i`th function body.

包含本模块中所有的函数体（定义），function section中的函数数量必须和code section中的函数体数量相同，并且顺序一致

| Field  | Type             | Description                              |
| ------ | ---------------- | ---------------------------------------- |
| count  | `varuint32`      | 函数体数量                                    |
| bodies | `function_body*` | sequence of [Function Bodies](http://webassembly.org/docs/binary-encoding/#function-bodies) |

### Data section

声明了被读入 linear memory中的已初始化的数据

| Field   | Type            | Description |
| ------- | --------------- | ----------- |
| count   | `varuint32`     | 数据段的数量      |
| entries | `data_segment*` | 如下          |

a `data_segment` is:

| Field  | Type        | Description                              |
| ------ | ----------- | ---------------------------------------- |
| index  | `varuint32` | the [linear memory index](http://webassembly.org/docs/modules/#linear-memory-index-space) (0 in the MVP) |
| offset | `init_expr` | an `i32` initializer expression that computes the offset at which to place the data |
| size   | `varuint32` | size of `data` (in bytes)                |
| data   | `bytes`     | sequence of `size` bytes                 |

### Name section

Custom section `name` field: `"name"`

name section是一个custom section，因此它被编码为id `0`，后面跟字符串”name“，与所有的custom section 一样，，该部分的格式错误并不会导致模块的验证失败，这依赖于它是如何实现对于畸形或者部分畸形的name section的处理机制。WebAssembly的实现为惰性读取和处理name section，在模块实例化后是否需要debugging（The WebAssembly implementation is also free to choose to read and process this section lazily, after the module has been instantiated, should debugging be required.）

name section应该只出现一次并且在Data section后，期望是，当在浏览器或其他开发环境中查看二进制WebAssembly模块时，name section中的数据将以 [text format](http://webassembly.org/docs/text-format/)的形式用作函数名和本地名。（The expectation is that, when a binary WebAssembly module is viewed in a browser or other development environment, the data in this section will be used as the names of functions and locals in the [text format](http://webassembly.org/docs/text-format/).）

name section包含一系列的name subsections

| Field             | Type        | Description                              |
| ----------------- | ----------- | ---------------------------------------- |
| name_type         | `varuint7`  | code identifying type of name contained in this subsection |
| name_payload_len  | `varuint32` | size of this subsection in bytes         |
| name_payload_data | `bytes`     | content of this section, of length `name_payload_len` |

因为name subsetcions的长度已被指定，未知的或不需要的subsections可以被引擎忽略。

当前有效的name_type为

| Name Type                                | Code | Description  |
| ---------------------------------------- | ---- | ------------ |
| [Module](http://webassembly.org/docs/binary-encoding/#module-name) | `0`  | 为模块指定名称      |
| [Function](http://webassembly.org/docs/binary-encoding/#function-names) | `1`  | 为函数指定名称      |
| [Local](http://webassembly.org/docs/binary-encoding/#local-names) | `2`  | 为函数的局部变量指定名称 |

当subsections 存在，它们必须已此顺序并最多出现一次，最后一个subsection的结尾必须与name section的最后一个字节一致.

#### Module name

module name subsection 为本模块指定一个名称，只包含一个字符串:

| Field    | Type        | Description                   |
| -------- | ----------- | ----------------------------- |
| name_len | `varuint32` | length of `name_str` in bytes |
| name_str | `bytes`     | UTF-8 encoding of the name    |

#### Name Map

下面的subsections , `name_map`为：

| Field | Type        | Description                          |
| ----- | ----------- | ------------------------------------ |
| count | `varuint32` | number of `naming` in names          |
| names | `naming*`   | sequence of `naming` sorted by index |

 `naming` 为:

| Field    | Type        | Description                    |
| -------- | ----------- | ------------------------------ |
| index    | `varuint32` | the index which is being named |
| name_len | `varuint32` | length of `name_str` in bytes  |
| name_str | `bytes`     | UTF-8 encoding of the name     |

#### Function names

function names subsection 是一个`name_map`， 它将名称分配给 [function index space](http://webassembly.org/docs/modules/#function-index-space) 的一个子集（导入和模块定义）

每个函数至多被命名一次，对一个函数的多次命名会导致该section变的畸形。

但是名称并不需要唯一，多个函数可以被命名为同一个名字，这在c++程序中很常见，其中包含二进制的多个编译单元可以包含具有相同名称的本地函数。

#### Local names

local names subsection是一个`name_map`， 它将名称分配给 [function index space](http://webassembly.org/docs/modules/#function-index-space) 的一个子集（导入和模块定义） ，给定函数的name_map将名称分配给局部变量索引的子集。

| Field | Type           | Description                              |
| ----- | -------------- | ---------------------------------------- |
| count | `varuint32`    | count of `local_names` in funcs          |
| funcs | `local_names*` | sequence of `local_names` sorted by index |

 `local_name` 为:

| Field     | Type        | Description                              |
| --------- | ----------- | ---------------------------------------- |
| index     | `varuint32` | the index of the function whose locals are being named |
| local_map | `name_map`  | assignment of names to local indices     |



# Function Bodies

Function bodies包含一系列的局部变量声明，后跟 [bytecode instructions](http://webassembly.org/docs/semantics/)，指令被编码为 [opcode](http://webassembly.org/docs/binary-encoding/#instruction-opcodes)，后跟0或者如下表，每个函数体必须已`end`opcode结束



| Field       | Type           | Description                              |
| ----------- | -------------- | ---------------------------------------- |
| body_size   | `varuint32`    | size of function body to follow, in bytes |
| local_count | `varuint32`    | number of local entries                  |
| locals      | `local_entry*` | local variables                          |
| code        | `byte*`        | bytecode of the function                 |
| end         | `byte`         | `0x0b`, indicating the end of the body   |

#### Local Entry

每个local entry 声明了知道类型的局部变量的数量，同一个类型可以有多个局部变量。

| Field | Type         | Description                              |
| ----- | ------------ | ---------------------------------------- |
| count | `varuint32`  | number of local variables of the following type |
| type  | `value_type` | type of the variables                    |

## Control flow operators 

| Name          | Opcode | Immediates                   | Description                              |                    |
| ------------- | ------ | ---------------------------- | ---------------------------------------- | ------------------ |
| `unreachable` | `0x00` |                              | trap immediately                         | 陷阱指令               |
| `nop`         | `0x01` |                              | no operation                             | 无操作                |
| `block`       | `0x02` | sig : `block_type`           | begin a sequence of expressions, yielding 0 or 1 values | 块开始，返回0或1个值        |
| `loop`        | `0x03` | sig : `block_type`           | begin a block which can also form control flow loops | 循环开始               |
| `if`          | `0x04` | sig : `block_type`           | begin if expression                      | if 开始              |
| `else`        | `0x05` |                              | begin else expression of if              |                    |
| `end`         | `0x0b` |                              | end a block, loop, or if                 | block ,loop,if 的结束 |
| `br`          | `0x0c` | relative_depth : `varuint32` | break that targets an outer nested block | 跳转到外部block目标       |
| `br_if`       | `0x0d` | relative_depth : `varuint32` | conditional break that targets an outer nested block | 条件跳转               |
| `br_table`    | `0x0e` | 如下                           | branch table control flow construct      |                    |
| `return`      | `0x0f` |                              | return zero or one value from this function | 返回0或1个值            |

`block`和`if`的sig字段，操作符指定了函数签名，它描述了它们对操作数堆栈的使用。(The *sig* fields of `block` and `if` operators specify function signatures which describe their use of the operand stack.)

`br_table` 运算符有一个直接操作数，编码如下:

| Field          | Type         | Description                              |                       |
| -------------- | ------------ | ---------------------------------------- | --------------------- |
| target_count   | `varuint32`  | number of entries in the target_table    | target_table中target数量 |
| target_table   | `varuint32*` | target entries that indicate an outer block or loop to which to break | 需跳转的外部block或loop的目标   |
| default_target | `varuint32`  | an outer block or loop to which to break in the default case | 默认目标                  |



`br_table` 操作符实现了一个间接分支，它接收一个可选的参数（同其他的分支）和一个额外的`i32`表达式作为输入，在`target_table`中给定偏移量上对block和loop的分支（branches to the block or loop at the given offset within the `target_table`），如果输入值越界，`br_table`将跳转到默认的目标`default_target`



## Call operators 

| Name            | Opcode | Immediates                               | Description                              |              |
| --------------- | ------ | ---------------------------------------- | ---------------------------------------- | ------------ |
| `call`          | `0x10` | function_index : `varuint32`             | call a function by its [index](http://webassembly.org/docs/modules/#function-index-space) | 根据index调用函数  |
| `call_indirect` | `0x11` | type_index : `varuint32`, reserved : `varuint1` | call a function indirect with an expected signature | 通过期望签名简介调用函数 |

`call_indirect`  操作符接受一个函数参数列表作为最后一个操作数和在表中索引（ as the last operand the index into the table），在MVP中必须为0。



## Parametric operators 

| Name     | Opcode | Immediates | Description                              |            |
| -------- | ------ | ---------- | ---------------------------------------- | ---------- |
| `drop`   | `0x1a` |            | ignore value                             | 忽略该值       |
| `select` | `0x1b` |            | select one of two values based on condition | 条件选择1个或2个值 |

## Variable access

| Name         | Opcode | Immediates                 | Description                              |                |
| ------------ | ------ | -------------------------- | ---------------------------------------- | -------------- |
| `get_local`  | `0x20` | local_index : `varuint32`  | read a local variable or parameter       | 读取一个局部变量或参数    |
| `set_local`  | `0x21` | local_index : `varuint32`  | write a local variable or parameter      | 写入一个局部变量或参数    |
| `tee_local`  | `0x22` | local_index : `varuint32`  | write a local variable or parameter and return the same value | 写入一个局部变量或参数并返回 |
| `get_global` | `0x23` | global_index : `varuint32` | read a global variable                   | 读取一个全局变量       |
| `set_global` | `0x24` | global_index : `varuint32` | write a global variable                  | 写入一个全局变量       |

## Memory-related operators ([described here](http://webassembly.org/docs/semantics/#linear-memory-accesses))

| Name             | Opcode | Immediate             | Description              |
| ---------------- | ------ | --------------------- | ------------------------ |
| `i32.load`       | `0x28` | `memory_immediate`    | load from memory         |
| `i64.load`       | `0x29` | `memory_immediate`    | load from memory         |
| `f32.load`       | `0x2a` | `memory_immediate`    | load from memory         |
| `f64.load`       | `0x2b` | `memory_immediate`    | load from memory         |
| `i32.load8_s`    | `0x2c` | `memory_immediate`    | load from memory         |
| `i32.load8_u`    | `0x2d` | `memory_immediate`    | load from memory         |
| `i32.load16_s`   | `0x2e` | `memory_immediate`    | load from memory         |
| `i32.load16_u`   | `0x2f` | `memory_immediate`    | load from memory         |
| `i64.load8_s`    | `0x30` | `memory_immediate`    | load from memory         |
| `i64.load8_u`    | `0x31` | `memory_immediate`    | load from memory         |
| `i64.load16_s`   | `0x32` | `memory_immediate`    | load from memory         |
| `i64.load16_u`   | `0x33` | `memory_immediate`    | load from memory         |
| `i64.load32_s`   | `0x34` | `memory_immediate`    | load from memory         |
| `i64.load32_u`   | `0x35` | `memory_immediate`    | load from memory         |
| `i32.store`      | `0x36` | `memory_immediate`    | store to memory          |
| `i64.store`      | `0x37` | `memory_immediate`    | store to memory          |
| `f32.store`      | `0x38` | `memory_immediate`    | store to memory          |
| `f64.store`      | `0x39` | `memory_immediate`    | store to memory          |
| `i32.store8`     | `0x3a` | `memory_immediate`    | store to memory          |
| `i32.store16`    | `0x3b` | `memory_immediate`    | store to memory          |
| `i64.store8`     | `0x3c` | `memory_immediate`    | store to memory          |
| `i64.store16`    | `0x3d` | `memory_immediate`    | store to memory          |
| `i64.store32`    | `0x3e` | `memory_immediate`    | store to memory          |
| `current_memory` | `0x3f` | reserved : `varuint1` | query the size of memory |
| `grow_memory`    | `0x40` | reserved : `varuint1` | grow the size of memory  |

`memory_immediate`如下:

| Name   | Type        | Description                              |                                 |
| ------ | ----------- | ---------------------------------------- | ------------------------------- |
| flags  | `varuint32` | a bitfield which currently contains the alignment in the least significant bits, encoded as `log2(alignment)` | 保存当前最不重要的位的位域，`log2(alignment)` |
| offset | `varuint32` | the value of the offset                  | 偏移量                             |

`alignment`必须位2的幂，作为额外的验证标准，alignment必须小于或等于`nature alignment`, `log(memory-access-size)`之后的最不重要的位必须设置位0.

`current_memory`和`grow_memory`在MVP中必须为0

## Constants 

| Name        | Opcode | Immediates         | Description                           |
| ----------- | ------ | ------------------ | ------------------------------------- |
| `i32.const` | `0x41` | value : `varint32` | a constant value interpreted as `i32` |
| `i64.const` | `0x42` | value : `varint64` | a constant value interpreted as `i64` |
| `f32.const` | `0x43` | value : `uint32`   | a constant value interpreted as `f32` |
| `f64.const` | `0x44` | value : `uint64`   | a constant value interpreted as `f64` |

## Comparison operators 

| Name       | Opcode | Immediate | Description |
| ---------- | ------ | --------- | ----------- |
| `i32.eqz`  | `0x45` |           | 等于0         |
| `i32.eq`   | `0x46` |           | 等于          |
| `i32.ne`   | `0x47` |           | 不等          |
| `i32.lt_s` | `0x48` |           | 小于 （有符号）    |
| `i32.lt_u` | `0x49` |           | 小于（无符号）     |
| `i32.gt_s` | `0x4a` |           | 大于 （有符号）    |
| `i32.gt_u` | `0x4b` |           | 大于（无符号）     |
| `i32.le_s` | `0x4c` |           | 小于或等于（有符号）  |
| `i32.le_u` | `0x4d` |           | 小于或等于（无符号）  |
| `i32.ge_s` | `0x4e` |           | 大于或等于（有符号）  |
| `i32.ge_u` | `0x4f` |           | 大于或等于（无符号）  |
| `i64.eqz`  | `0x50` |           | 等于0         |
| `i64.eq`   | `0x51` |           | 等于          |
| `i64.ne`   | `0x52` |           | 不等          |
| `i64.lt_s` | `0x53` |           | 小于 （有符号）    |
| `i64.lt_u` | `0x54` |           | 小于（无符号）     |
| `i64.gt_s` | `0x55` |           | 大于 （有符号）    |
| `i64.gt_u` | `0x56` |           | 大于（无符号）     |
| `i64.le_s` | `0x57` |           | 小于或等于（有符号）  |
| `i64.le_u` | `0x58` |           | 小于或等于（无符号）  |
| `i64.ge_s` | `0x59` |           | 大于或等于（有符号）  |
| `i64.ge_u` | `0x5a` |           | 大于或等于（无符号）  |
| `f32.eq`   | `0x5b` |           | 等于          |
| `f32.ne`   | `0x5c` |           | 不等          |
| `f32.lt`   | `0x5d` |           | 小于          |
| `f32.gt`   | `0x5e` |           | 大于          |
| `f32.le`   | `0x5f` |           | 小于或等于       |
| `f32.ge`   | `0x60` |           | 大于或等于       |
| `f64.eq`   | `0x61` |           | 等于          |
| `f64.ne`   | `0x62` |           | 不等          |
| `f64.lt`   | `0x63` |           | 小于          |
| `f64.gt`   | `0x64` |           | 大于          |
| `f64.le`   | `0x65` |           | 小于或等于       |
| `f64.ge`   | `0x66` |           | 大于或等于       |

## Numeric operators 

| Name           | Opcode | Immediate | Description                              |
| -------------- | ------ | --------- | ---------------------------------------- |
| `i32.clz`      | `0x67` |           | 零基数指令，用于计算最高符号位与第一个1之间的0的个数              |
| `i32.ctz`      | `0x68` |           | 计算尾部0的个数                                 |
| `i32.popcnt`   | `0x69` |           | 位1计数，统计有多少个“为1的位”                        |
| `i32.add`      | `0x6a` |           | 符号未知 加                                   |
| `i32.sub`      | `0x6b` |           | 符号未知 减                                   |
| `i32.mul`      | `0x6c` |           | 符号未知 乘                                   |
| `i32.div_s`    | `0x6d` |           | 有符号除，结果被截断为0(result is truncated toward zero) |
| `i32.div_u`    | `0x6e` |           | 无符号除，结果向上取整(floor)                       |
| `i32.rem_s`    | `0x6f` |           | 有符号取余（符号不同时，结果符号同第一个变量）                  |
| `i32.rem_u`    | `0x70` |           | 无符号取余                                    |
| `i32.and`      | `0x71` |           | 位与                                       |
| `i32.or`       | `0x72` |           | 位或                                       |
| `i32.xor`      | `0x73` |           | 位异或                                      |
| `i32.shl`      | `0x74` |           | 左移位                                      |
| `i32.shr_s`    | `0x75` |           | 逻辑右移位（左补符号）                              |
| `i32.shr_u`    | `0x76` |           | 算法右移位（左补0）                               |
| `i32.rotl`     | `0x77` |           | 循环左移                                     |
| `i32.rotr`     | `0x78` |           | 循环右移                                     |
| `i64.clz`      | `0x79` |           | 零基数指令，用于计算最高符号位与第一个1之间的0的个数              |
| `i64.ctz`      | `0x7a` |           | 计算尾部0的个数                                 |
| `i64.popcnt`   | `0x7b` |           | 位1计数，统计有多少个“为1的位”                        |
| `i64.add`      | `0x7c` |           | 符号未知 加                                   |
| `i64.sub`      | `0x7d` |           | 符号未知 减                                   |
| `i64.mul`      | `0x7e` |           | 符号未知 乘                                   |
| `i64.div_s`    | `0x7f` |           | 有符号除，结果被截断为0(result is truncated toward zero) |
| `i64.div_u`    | `0x80` |           | 无符号除，结果向上取整(floor)                       |
| `i64.rem_s`    | `0x81` |           | 有符号取余（符号不同时，结果符号同第一个变量）                  |
| `i64.rem_u`    | `0x82` |           | 无符号取余                                    |
| `i64.and`      | `0x83` |           | 位与                                       |
| `i64.or`       | `0x84` |           | 位或                                       |
| `i64.xor`      | `0x85` |           | 位异或                                      |
| `i64.shl`      | `0x86` |           | 左移位                                      |
| `i64.shr_s`    | `0x87` |           | 逻辑右移位（左补符号）                              |
| `i64.shr_u`    | `0x88` |           | 算法右移位（左补0）                               |
| `i64.rotl`     | `0x89` |           | 循环左移                                     |
| `i64.rotr`     | `0x8a` |           | 循环右移                                     |
| `f32.abs`      | `0x8b` |           | 绝对值                                      |
| `f32.neg`      | `0x8c` |           | 取反                                       |
| `f32.ceil`     | `0x8d` |           | 向下取整                                     |
| `f32.floor`    | `0x8e` |           | 向上取整                                     |
| `f32.trunc`    | `0x8f` |           | 四舍五入到接近零的整数                              |
| `f32.nearest`  | `0x90` |           | 四舍五入到接近的整数                               |
| `f32.sqrt`     | `0x91` |           | 平方根                                      |
| `f32.add`      | `0x92` |           | 加                                        |
| `f32.sub`      | `0x93` |           | 减                                        |
| `f32.mul`      | `0x94` |           | 乘                                        |
| `f32.div`      | `0x95` |           | 除                                        |
| `f32.min`      | `0x96` |           | 取最小，如果有NaN，则返回NaN                        |
| `f32.max`      | `0x97` |           | 取最大，如果有NaN，则返回NaN                        |
| `f32.copysign` | `0x98` |           | 将第二个参数的符号赋予第一个参数                         |
| `f64.abs`      | `0x99` |           | 绝对值                                      |
| `f64.neg`      | `0x9a` |           | 取反                                       |
| `f64.ceil`     | `0x9b` |           | 向下取整                                     |
| `f64.floor`    | `0x9c` |           | 向上取整                                     |
| `f64.trunc`    | `0x9d` |           | 四舍五入到接近整数，若接近0，则取0                       |
| `f64.nearest`  | `0x9e` |           | 四舍五入到接近的偶数                               |
| `f64.sqrt`     | `0x9f` |           | 平方根                                      |
| `f64.add`      | `0xa0` |           | 加                                        |
| `f64.sub`      | `0xa1` |           | 减                                        |
| `f64.mul`      | `0xa2` |           | 乘                                        |
| `f64.div`      | `0xa3` |           | 除                                        |
| `f64.min`      | `0xa4` |           | 取最小，如果有NaN，则返回NaN                        |
| `f64.max`      | `0xa5` |           | 取最大，如果有NaN，则返回NaN                        |
| `f64.copysign` | `0xa6` |           | 将第二个参数的符号赋予第一个参数                         |

## Conversions 

| Name                | Opcode | Immediate | Description              |
| ------------------- | ------ | --------- | ------------------------ |
| `i32.wrap/i64`      | `0xa7` |           | 将64bit的整数转位32bit整数       |
| `i32.trunc_s/f32`   | `0xa8` |           | 将32bit的浮点数转换为32bit的有符号整数 |
| `i32.trunc_u/f32`   | `0xa9` |           | 将32bit的浮点数转换为32bit的无符号整数 |
| `i32.trunc_s/f64`   | `0xaa` |           | 将64bit的浮点数转换为32bit的有符号整数 |
| `i32.trunc_u/f64`   | `0xab` |           | 将64bit的浮点数转换为32bit的无符号整数 |
| `i64.extend_s/i32`  | `0xac` |           | 将32bit的整数扩展为64bit有符号整数   |
| `i64.extend_u/i32`  | `0xad` |           | 将32bit的整数扩展为64bit无符号整数   |
| `i64.trunc_s/f32`   | `0xae` |           | 将32bit的浮点数转换为64bit的有符号整数 |
| `i64.trunc_u/f32`   | `0xaf` |           | 将32bit的浮点数转换为64bit的无符号整数 |
| `i64.trunc_s/f64`   | `0xb0` |           | 将64bit的浮点数转换为64bit的有符号整数 |
| `i64.trunc_u/f64`   | `0xb1` |           | 将64bit的浮点数转换为64bit的无符号整数 |
| `f32.convert_s/i32` | `0xb2` |           | 将32bit的整数转换为32bit的有符号浮点数 |
| `f32.convert_u/i32` | `0xb3` |           | 将32bit的整数转换为32bit的无符号浮点数 |
| `f32.convert_s/i64` | `0xb4` |           | 将64bit的整数转换为32bit的有符号浮点数 |
| `f32.convert_u/i64` | `0xb5` |           | 将64bit的整数转换为32bit的无符号浮点数 |
| `f32.demote/f64`    | `0xb6` |           | 将64bit的浮点数降级为32bit的浮点数   |
| `f64.convert_s/i32` | `0xb7` |           | 将32bit的整数转换为64bit的有符号整数  |
| `f64.convert_u/i32` | `0xb8` |           | 将32bit的整数转换为64bit的无符号整数  |
| `f64.convert_s/i64` | `0xb9` |           | 将64bit的整数转换为64bit的有符号整数  |
| `f64.convert_u/i64` | `0xba` |           | 将64bit的整数转换为64bit的无符号整数  |
| `f64.promote/f32`   | `0xbb` |           | 将32bit的浮点数升级为64bit的有符号整数 |

## Reinterpretations 

| Name                  | Opcode | Immediate | Description         |
| --------------------- | ------ | --------- | ------------------- |
| `i32.reinterpret/f32` | `0xbc` |           | 将32bit浮点数强转为32bit整数 |
| `i64.reinterpret/f64` | `0xbd` |           | 将64bit浮点数强转为64bit整数 |
| `f32.reinterpret/i32` | `0xbe` |           | 将32bit整数强转为32bit浮点  |
| `f64.reinterpret/i64` | `0xbf` |           | 将64bit整数强转为64bit浮点数 |
