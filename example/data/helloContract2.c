char* hello(char* name) {
  return strconcat("hello ",name);
}
char* invoke(char* method,char* args) {
  if (strcmp(method, "hello")==0){
        char* name = ReadStringParam(args);
        char* res = hello(name);
        // char* result = JsonMashalResult(res,"string",1);
        return res;
    }
  return "fail";
}