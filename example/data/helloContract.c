char* strconcat (char* a,char* b) {
  int lena = arrayLen(a);
  int lenb = arrayLen(b);
  char* r = (char*) malloc((lena+lenb)*sizeof(char));
  for (int i=0;i<lena;i++){
    r[i]=a[i];
  }
  for (int i=0;i<lenb;i++){
    r[lena+i]=b[i];
  }
  return r;
}

char* hello(char* name) {
  return strconcat("hello ",name);
}
char* invoke(char* method,char* args) {
  if (strcmp(method ,"init")==0 ){
        return "init success!";
  }
  if (strcmp(method, "hello")==0){
        char* name = ReadStringParam(args);
        char* res = hello(name);
        char* result = JsonMashalResult(res,"string",1);
        return result;
    }
  return "fail";
}