#include "extensions.h"
#include "rocksdb/c.h"
#include <stdio.h>
#include <stdlib.h>

// gcc -O0 -g -I/home/me/rocksdb/include/ -L/home/me/rocksdb test.c -lrocksdb -lbz2 -lpthread -lsnappy -lz -lextensions

int main() {
  char* errptr = NULL;
  rocksdb_options_t* options = rocksdb_options_create();
  const rocksdb_options_t** cf_options = (const rocksdb_options_t**) calloc(5, sizeof(rocksdb_options_t*));
  char **cfs= (char**)calloc(5, sizeof(char*));
  int i = 0;

  for (i=0;i<4;i++) {
    cfs[i]=(char*)calloc(2, sizeof(char));
    cfs[i][0]='0'+i;
    cfs[i][1]='\0';
    cf_options[i]= rocksdb_options_create();
  }
  cfs[4]=(char*)calloc(8, sizeof(char));
  cf_options[4]= rocksdb_options_create();
  sprintf(cfs[4],"%s","default");
  cfs[4][7]='\0';


  rocksdb_column_family_handle_t** cf_handles = (rocksdb_column_family_handle_t**) calloc(5, sizeof(rocksdb_column_family_handle_t*));
  rocksdb_t* db = rocksdb_open_column_families(options, "/tmp/multicf_db/", 5, (const char**)cfs, cf_options, cf_handles, &errptr);
  if (errptr!=NULL) {
    printf("OPENING: %s\n", errptr);
    return -1;
  }

  rocksdb_iterator_t** iters = (rocksdb_iterator_t**) calloc(5, sizeof(rocksdb_iterator_t*));
  get_iterators(rocksdb_readoptions_create(), db, cf_handles, iters, 5, &errptr);
  if (errptr!=NULL) {
    printf("ITERATORS: %s\n", errptr);
    return -1;
  }

  for (i=0;i<5;i++) {
    printf("FOR: %d\n", i);
    rocksdb_iter_seek_to_first(iters[i]);
    while (rocksdb_iter_valid(iters[i])) {
      size_t size;
      const char* data = rocksdb_iter_key(iters[i], &size);
      printf("%.*s\n", (int)size, data);
      rocksdb_iter_next(iters[i]);
    }
  }
  return 0;
}
