//==============================================================================
//
// Minunit
//
//==============================================================================

#define mu_fail(MSG, ...) do {\
    fprintf(stderr, "%s:%d: " MSG "\n", __FILE__, __LINE__, ##__VA_ARGS__);\
    return 1;\
} while(0)

#define mu_assert(TEST, MSG, ...) do {\
    if (!(TEST)) {\
        fprintf(stderr, "%s:%d: %s " MSG "\n", __FILE__, __LINE__, #TEST, ##__VA_ARGS__);\
        return 1;\
    }\
} while (0)

#define mu_run_test(TEST) do {\
    int rc = TEST();\
    tests_run++; \
    if (rc) {\
        return rc;\
    }\
} while (0)

#define RUN_TESTS() int main() {\
   int rc = all_tests();\
   if(rc == 0) {\
       printf("ALL TESTS PASSED\n");\
   }\
   printf("Tests run: %d\n", tests_run);\
   return rc;\
}

int tests_run;



//==============================================================================
//
// Constants
//
//==============================================================================

// The temporary file used for file operations in the test suite.
#define TEMPFILE "/tmp/skytemp"


//==============================================================================
//
// Helpers
//
//==============================================================================

// Copy a database from the fixtures directory into the tmp/db directory.
#define copydb(DB) \
    char _copydb_cmd[1024]; \
    snprintf(_copydb_cmd, 1024, "tests/copydb.sh %s", DB); \
    system(_copydb_cmd)
    
// Asserts that a block has a specific block id and object id range.
#define mu_assert_block_info(INDEX, ID, MIN_OBJECT_ID, MAX_OBJECT_ID, MIN_TIMESTAMP, MAX_TIMESTAMP, SPANNED) \
    mu_assert(object_file->infos[INDEX].id == ID, "Block " #INDEX " id expected to be " #ID); \
    mu_assert(object_file->infos[INDEX].min_object_id == MIN_OBJECT_ID, "Block " #INDEX " min object id expected to be " #MIN_OBJECT_ID); \
    mu_assert(object_file->infos[INDEX].max_object_id == MAX_OBJECT_ID, "Block " #INDEX " max object id expected to be " #MAX_OBJECT_ID); \
    mu_assert(object_file->infos[INDEX].min_timestamp == MIN_TIMESTAMP, "Block " #INDEX " min timestamp expected to be " #MIN_TIMESTAMP); \
    mu_assert(object_file->infos[INDEX].max_timestamp == MAX_TIMESTAMP, "Block " #INDEX " max timestamp expected to be " #MAX_TIMESTAMP); \
    mu_assert(object_file->infos[INDEX].spanned == SPANNED, "Block " #INDEX " spanned expected to be " #SPANNED);

// Asserts action data.
#define mu_assert_action(INDEX, ID, NAME) \
    mu_assert(object_file->actions[INDEX].id == ID, "Expected action #" #INDEX " id to be: " #ID); \
    mu_assert(biseqcstr(object_file->actions[INDEX].name, NAME) == 1, "Expected action #" #INDEX " name to be: " #NAME);

// Asserts property data.
#define mu_assert_property(INDEX, ID, NAME) \
    mu_assert(object_file->properties[INDEX].id == ID, "Expected property #" #INDEX " id to be: " #ID); \
    mu_assert(biseqcstr(object_file->properties[INDEX].name, NAME) == 1, "Expected property #" #INDEX " name to be: " #NAME);

// Asserts the contents of the temp file.
#define mu_assert_tempfile(EXP_FILENAME) \
    unsigned char tempch; \
    unsigned char expch; \
    FILE *tempfile = fopen(TEMPFILE, "r"); \
    FILE *expfile = fopen(EXP_FILENAME, "r"); \
    if(tempfile == NULL) mu_fail("Cannot open tempfile: %s", TEMPFILE); \
    if(expfile == NULL) mu_fail("Cannot open expectation file: %s", EXP_FILENAME); \
    while(1) { \
        fread(&expch, 1, 1, expfile); \
        fread(&tempch, 1, 1, tempfile); \
        if(feof(expfile) || feof(tempfile)) break; \
        if(tempch != expch) { \
            mu_fail("Expected 0x%02x, received 0x%02x at location %ld in tempfile", expch, tempch, (ftell(expfile)-1)); \
        } \
    } \
    if(!feof(tempfile)) mu_fail("Tempfile length longer than expected: %s", TEMPFILE); \
    if(!feof(expfile)) mu_fail("Tempfile length shorter than expected: %s", EXP_FILENAME); \
    fclose(tempfile); \
    fclose(expfile);
