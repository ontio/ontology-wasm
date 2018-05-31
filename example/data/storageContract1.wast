(module
 (type $FUNCSIG$i (func (result i32)))
 (type $FUNCSIG$ii (func (param i32) (result i32)))
 (type $FUNCSIG$vii (func (param i32 i32)))
 (type $FUNCSIG$iii (func (param i32 i32) (result i32)))
 (import "env" "JsonMashalResult" (func $JsonMashalResult (param i32 i32) (result i32)))
 (import "env" "ReadStringParam" (func $ReadStringParam (param i32) (result i32)))
 (import "env" "getStorage" (func $getStorage (param i32) (result i32)))
 (import "env" "putStorage" (func $putStorage (param i32 i32)))
 (import "env" "strcmp" (func $strcmp (param i32 i32) (result i32)))
 (import "env" "strconcat" (func $strconcat (param i32 i32) (result i32)))
 (table 0 anyfunc)
 (memory $0 1)
 (data (i32.const 16) "hello \00")
 (data (i32.const 32) "put success\00")
 (data (i32.const 48) "hello\00")
 (data (i32.const 64) "string\00")
 (data (i32.const 80) "get\00")
 (data (i32.const 96) "put\00")
 (data (i32.const 112) "unsupport method\00")
 (export "memory" (memory $0))
 (export "hello" (func $hello))
 (export "get" (func $get))
 (export "put" (func $put))
 (export "invoke" (func $invoke))
 (func $hello (; 6 ;) (param $0 i32) (result i32)
  (call $strconcat
   (i32.const 16)
   (get_local $0)
  )
 )
 (func $get (; 7 ;) (param $0 i32) (result i32)
  (call $getStorage
   (get_local $0)
  )
 )
 (func $put (; 8 ;) (param $0 i32) (param $1 i32) (result i32)
  (call $putStorage
   (get_local $0)
   (get_local $1)
  )
  (i32.const 32)
 )
 (func $invoke (; 9 ;) (param $0 i32) (param $1 i32) (result i32)
  (local $2 i32)
  (block $label$0
   (block $label$1
    (block $label$2
     (br_if $label$2
      (i32.eqz
       (call $strcmp
        (get_local $0)
        (i32.const 48)
       )
      )
     )
     (br_if $label$1
      (i32.eqz
       (call $strcmp
        (get_local $0)
        (i32.const 80)
       )
      )
     )
     (set_local $2
      (i32.const 112)
     )
     (br_if $label$0
      (call $strcmp
       (get_local $0)
       (i32.const 96)
      )
     )
     (call $putStorage
      (tee_local $0
       (call $ReadStringParam
        (get_local $1)
       )
      )
      (call $strconcat
       (i32.const 16)
       (get_local $0)
      )
     )
     (set_local $2
      (i32.const 32)
     )
     (br $label$0)
    )
    (set_local $2
     (call $strconcat
      (i32.const 16)
      (call $ReadStringParam
       (get_local $1)
      )
     )
    )
    (br $label$0)
   )
   (set_local $2
    (call $getStorage
     (call $ReadStringParam
      (get_local $1)
     )
    )
   )
  )
  (call $JsonMashalResult
   (get_local $2)
   (i32.const 64)
  )
 )
)
