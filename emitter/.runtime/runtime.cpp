#include <stdlib.h>
#include <stdio.h>
#include <string>
#include <vector>
#include <algorithm>

// TODO: wow this is awful.
#define TS_OBJECT_FIELD_NUM 10

// TODO: don't do this.
#define UNDEFINED 0xDEADBEEF

enum CoreType
{
	not = 0,
	CoreTypeTsNum = 1,
	CoreTypeTsString = 2,
	CoreTypeTsFunction = 3,
};

template <typename T>
class IntrinsicTsObject : TsObject
{
public:
	T value;

	IntrinsicTsObject(T value) : value(value) {}

	bool operator==(const IntrinsicTsObject<T> &other) const
	{
		return value == other.value;
	}
};

class TsNum : TsObject
{
public:
	TsNum(int num) : TsObject()
	{
		addIntrinsicField<int>("value", num);
	}

	bool operator==(const TsString &other) const
	{
		return getField("value")->value == getField("value")->value;
	}
};

class TsString : TsObject
{
public:
	TsString(std::string value)
	{
		addIntrinsicField<std::string>("value", value);
	}

	bool operator==(const TsString &other) const
	{
		return getField("value")->value == getField("value")->value;
	}
};

class TsFunctionParam
{
	std::string name;
	int type_id;
};

class TsFunction : TsObject
{
public:
	int num_params;
	std::vector<TsFunctionParam> params;
};

class TsObjectFieldDescriptor
{
public:
	TsString name;

	TsObjectFieldDescriptor(TsString name)
		: name(name)
	{
	}
};

class TsObjectField
{
public:
	TsObjectFieldDescriptor descriptor;
	std::shared_ptr<TsObject> value;

	TsObjectField(TsObjectFieldDescriptor descriptor) : TsObjectField(descriptor, NULL) {}

	TsObjectField(TsObjectFieldDescriptor descriptor, std::shared_ptr<TsObject> value)
		: descriptor(std::move(descriptor)), value(value)
	{
	}
};

class TsObject
{
public:
	std::vector<std::shared_ptr<TsObjectField>> fields;

	CoreType core_type;

	union
	{
		TsNum *num;
		TsString *str;
		TsFunction *func;
	} core_type_value;

	TsObject() : TsObject(std::vector<std::shared_ptr<TsObjectField>>()) {}

	TsObject(std::vector<std::shared_ptr<TsObjectField>> fields)
		: fields(std::move(fields)) {}

	std::shared_ptr<TsObjectField> getField(const std::string &field_name) const
	{
		return *std::find_if(fields.begin(), fields.end(), [](auto field)
							 { return *field->descriptor->name == field_name });
	}

	TsObjectFieldDescriptor getFieldDescriptor(const std::string &field_name)
	{
		return getField(field_name)->descriptor;
	}

	void setField(const std::string &field_name, std::shared_ptr<TsObject> value)
	{
		auto field = getField(field_name);

		// If there's no existing field, create one.
		if (field == NULL)
		{
			// Create a new field.
			auto new_field = std::make_shared<TsObjectField>(TsObjectField(TsObjectFieldDescriptor(field_name), value));

			// Attach the field.
			fields.push_back(new_field);

			return;
		}

		// Otherwise, just set the field.
		// TODO: handle freeing memory (whoops!).
		field->value = value;
	}

	template <typename T>
	T addIntrinsicField(const std::string &fieldName, T value)
	{
		auto descriptor = TsObjectFieldDescriptor(TsString(fieldName));
		auto object = IntrinsicTsObject(value);
		auto field = TsObjectField(descriptor, std::make_shared<TsObject>(object));

		fields.push_back(std::make_shared<TsObjectField>(field));
	}

	template <typename T>
	T getIntrinsicField(const std::string &fieldName)
	{
		auto field = (IntrinsicTsObject<T>)getField(fieldName);

		return field.value;
	}
};

void ts_main();

int main()
{
	ts_main();

	return 0;
}
