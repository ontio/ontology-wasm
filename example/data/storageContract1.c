void putStorage(char* key,char* value);
char* getStorage(char* key);
void log(char* v);


char* hello(char* name) {
  return strconcat("hello ",name);
}
char* get(char* name) {
  return getStorage(name);
}
char* put(char* name,char* value) {
  putStorage(name,value);
  return "put success";
}


char* invoke(char* method,char* args) {
  if (strcmp(method, "hello")==0){
        char* name = ReadStringParam(args);
        char* res = hello(name);
        char* result = JsonMashalResult(res,"string");
        return result;
    }
    if (strcmp(method, "get")==0){
        char* name = ReadStringParam(args);
        char* res = get(name);
        char* result = JsonMashalResult(res,"string");
        return result;
    }
    if (strcmp(method, "put")==0){
        char* name = ReadStringParam(args);
        char* v = hello(name);
        char* res = put(name,v);
        char* result = JsonMashalResult(res,"string");
        return result;
    }
  return JsonMashalResult("unsupport method","string");
}