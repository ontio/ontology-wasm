(module
 (type $FUNCSIG$i (func (result i32)))
 (type $FUNCSIG$iii (func (param i32 i32) (result i32)))
 (type $FUNCSIG$ii (func (param i32) (result i32)))
 (import "env" "ReadStringParam" (func $ReadStringParam (param i32) (result i32)))
 (import "env" "strcmp" (func $strcmp (param i32 i32) (result i32)))
 (import "env" "strconcat" (func $strconcat (param i32 i32) (result i32)))
 (table 0 anyfunc)
 (memory $0 1)
 (data (i32.const 16) "hello \00")
 (data (i32.const 32) "hello\00")
 (data (i32.const 48) "fail\00")
 (export "memory" (memory $0))
 (export "hello" (func $hello))
 (export "invoke" (func $invoke))
 (func $hello (; 3 ;) (param $0 i32) (result i32)
  (call $strconcat
   (i32.const 16)
   (get_local $0)
  )
 )
 (func $invoke (; 4 ;) (param $0 i32) (param $1 i32) (result i32)
  (block $label$0
   (br_if $label$0
    (i32.eqz
     (call $strcmp
      (get_local $0)
      (i32.const 32)
     )
    )
   )
   (return
    (i32.const 48)
   )
  )
  (call $strconcat
   (i32.const 16)
   (call $ReadStringParam
    (get_local $1)
   )
  )
 )
)
