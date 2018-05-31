int arrayLen(void *a);
void * malloc(int size);
char * strconcat(char * a, char * b){
	int lena = arrayLen(a);
	int lenb = arrayLen(b);
	char * res = (char *)malloc((lena + lenb)*sizeof(char));
	for (int i = 0 ;i < lena ;i++){
		res[i] = a[i];
	}

	for (int j = 0; j < lenb ;j++){
		res[lena + j] = b[j];
	}
	return res;
}
char* hello(char* name) {
  return strconcat("hello ",name);
}
