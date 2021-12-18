#include <stdlib.h>
#include <string.h>
#include <stdio.h>

// TODO: wow this is awful.
#define TS_OBJECT_FIELD_NUM 10

// TODO: don't do this.
#define UNDEFINED 0xDEADBEEF

typedef enum core_type
{
	not = 0,
	ts_num = 1,
	ts_string = 2,
	ts_function = 3,
} core_type;

typedef struct ts_num
{
	int value;
} ts_num;

typedef struct ts_string
{
	char *value;
	int len;
} ts_string;

typedef struct ts_function_param
{
	char *name;
	int type_id;
} ts_function_param;

typedef struct ts_function
{
	int num_params;
	ts_function_param params[];
} ts_function;

typedef struct ts_object_field_descriptor
{
	ts_string *name;
	int type_id;
	void *metadata;
} ts_object_field_descriptor;

typedef struct ts_object_field
{
	ts_object_field_descriptor *descriptor;
	void *value;
} ts_object_field;

typedef struct ts_object
{
	ts_object_field *fields[TS_OBJECT_FIELD_NUM];

	core_type core_type;

	union
	{
		ts_num *num;
		ts_string *str;
		ts_function *func;
	} core_type_value;

	int next_field_index;
} ts_object;

int ts_object_get_field_number(ts_object* obj, ts_string* field_name) {
	for (int i = 0; i < TS_OBJECT_FIELD_NUM; i++) {
		ts_object_field* field = obj->fields[i];
		if (field == NULL) {
			continue;
		}
		
		ts_object_field_descriptor* descriptor = field->descriptor;
		if (strcmp(descriptor->name->value, field_name->value) == 0) {
			return i;
		}
	}

	return -1;
}

void* ts_object_get_field(ts_object* obj, ts_string* name) {
	int field_number = ts_object_get_field_number(obj, name);
	if (field_number == -1) {
		return UNDEFINED;
	}

	return obj->fields[field_number]->value;
}

ts_object_field_descriptor* ts_object_get_field_descriptor(ts_object* obj, ts_string* field_name) {
	return obj->fields[ts_object_get_field_number(obj, field_name)]->descriptor;
}

void ts_object_set_field(ts_object* obj, ts_string* field_name, void* value) {
	ts_object_field_descriptor* field = ts_object_get_field_descriptor(obj, field_name);
	int field_number = ts_object_get_field_number(obj, field_name);

	// If there's no existing field, create one.
	if (field_number == -1) {
		int new_field_number = obj->next_field_index++;

		// Create a new field.
		ts_object_field* new_field = (ts_object_field*)malloc(sizeof(ts_object_field));
		new_field->descriptor = field;
		new_field->value = value;

		// Attach the field.
		obj->fields[new_field_number] = new_field;

		return;
	}

	// Otherwise, just set the field.
	// TODO: handle freeing memory (whoops!).
	obj->fields[field_number]->value = value;
}

ts_object* ts_object_new(ts_object_field* fields[], int n) {
	ts_object* obj = (ts_object*)malloc(sizeof(ts_object));
	obj->next_field_index = 0;

	for (int i = 0; i < n; i++) {
		int field_index = obj->next_field_index++;

		obj->fields[field_index] = fields[i];
	}

	return obj;
}

ts_object_field* ts_object_field_new(ts_object_field_descriptor* descriptor, void* value) {
	ts_object_field* field = (ts_object_field*)malloc(sizeof(ts_object_field));
	field->descriptor = descriptor;
	field->value = value;

	return field;
}

ts_object_field_descriptor* ts_object_field_descriptor_new(ts_string* name, int type_id, void* metadata) {
	ts_object_field_descriptor* descriptor = (ts_object_field_descriptor*)malloc(sizeof(ts_object_field_descriptor));
	descriptor->name = name;
	descriptor->type_id = type_id;
	descriptor->metadata = metadata;

	return descriptor;
}

ts_num* ts_num_new(int num) {
	ts_num* result = (ts_num*)malloc(sizeof(ts_num));
	result->value = num;

	return result;
}

ts_string* ts_string_new(char* str) {
	ts_string* result = (ts_string*)malloc(sizeof(ts_string));
	result->value = str;
	result->len = strlen(str);

	return result;
}

void ts_Console_log(ts_string* str) {
	printf("%s\n", str->value);
}

typedef struct ts_Console {
	void (*log)(ts_string* str);
} ts_Console;

ts_Console* console;

static void init_globals() {
	console = (ts_Console*)malloc(sizeof(ts_Console));
	console->log = ts_Console_log;
}

void ts_main();

int main() {
	init_globals();

	ts_main();

	return 0;
}
ts_num* ts_foo(ts_string* a, ts_num* b) {
	ts_num* ay = ts_num_new(5);
	ts_string* bee = ts_string_new("bar");
	return ay;
}

ts_string* ts_blah() {
	ts_string* foo = ts_string_new("asdf");
	ts_Console* bar = console;
	bar = console;
	return foo;
}

void ts_blah2() {
}

void ts_main() {
	ts_object* obj = ts_object_new((ts_object_field*[]){ts_object_field_new(ts_object_field_descriptor_new(ts_string_new("foo"), 0, NULL), ts_string_new("bar")), ts_object_field_new(ts_object_field_descriptor_new(ts_string_new("baz"), 0, NULL), ts_num_new(5)), ts_object_field_new(ts_object_field_descriptor_new(ts_string_new("qux"), 0, NULL), ts_object_new((ts_object_field*[]){ts_object_field_new(ts_object_field_descriptor_new(ts_string_new("a"), 0, NULL), ts_string_new("a"))}, 1))}, 3);
	ts_object_set_field((ts_object*)ts_object_get_field(obj, ts_string_new("qux")), ts_string_new("a"), ts_string_new("b"));
	ts_blah();
}


