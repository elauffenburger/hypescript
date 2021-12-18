#include <stdlib.h>
#include <stdio.h>
#include <string>
#include <vector>
#include <algorithm>

// TODO: don't do this.
#define UNDEFINED 0xDEADBEEF

class TsCoreHelpers
{
public:
	template <typename T>
	static std::vector<T> toVector(T value...)
	{
		auto result = std::vector<T>();

		va_list args;
		va_start(args, value);

		while (*value)
		{
			result.push_back(*value);

			++value;
		}

		va_end(args);

		return result;
	}

	template <typename T>
	static std::vector<T> toVector()
	{
		return std::vector<T>();
	}
};

enum TypeId
{
	TypeIdNone = 0,
	TypeIdTsObject = 1,
	TypeIdTsNum = 2,
	TypeIdTsString = 3,
	TypeIdTsFunction = 4,
	TypeIdVoid = 5,
};

template <typename T>
class IntrinsicTsObject : TsObject
{
public:
	T value;

	IntrinsicTsObject(T value)
		: value(value) {}

	bool operator==(const IntrinsicTsObject<T> &other) const
	{
		return value == other.value;
	}
};

class TsNum : TsObject
{
public:
	TsNum(int num)
		: TsObject(TypeIdTsNum)
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
		: TsObject(TypeIdTsString)
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
public:
	std::string name;
	int type_id;

	TsFunctionParam(std::string name, int type_id)
		: name(std::move(name)),
		  type_id(std::move(type_id)) {}
};

class TsFunctionArg
{
public:
	std::string name;
	std::shared_ptr<TsObject> value;

	TsFunctionArg(std::string name, std::shared_ptr<TsObject> value)
		: name(std::move(name)),
		  value(value) {}

	static const TsFunctionArg &findArg(const std::vector<TsFunctionArg> &args, const std::string &argName)
	{
		return *std::find_if(args.begin(), args.end(), [](auto arg)
							 { return arg.name == argName; });
	}
};

typedef std::shared_ptr<TsObject> (*TsFunctionFn)(std::vector<TsFunctionArg> args);

class TsFunction : TsObject
{
public:
	std::string name;
	std::vector<TsFunctionParam> params;

	TsFunctionFn fn;

	TsFunction(std::string name, std::vector<TsFunctionParam> params, TsFunctionFn fn)
		: name(std::move(name)),
		  params(std::move(params)),
		  fn(fn),
		  TsObject(TypeIdTsFunction) {}

	std::shared_ptr<TsObject> invoke(std::vector<TsFunctionArg> args) override
	{
		return fn(args);
	}
};

class TsObjectFieldDescriptor
{
public:
	TsString name;
	int typeId;

	TsObjectFieldDescriptor(TsString name, int typeId)
		: name(std::move(name)),
		  typeId(typeId)
	{
	}
};

class TsObjectField
{
public:
	TsObjectFieldDescriptor descriptor;
	std::shared_ptr<TsObject> value;

	TsObjectField(TsObjectFieldDescriptor descriptor)
		: TsObjectField(descriptor, NULL) {}

	TsObjectField(TsObjectFieldDescriptor descriptor, std::shared_ptr<TsObject> value)
		: descriptor(std::move(descriptor)),
		  value(value) {}
};

class TsObject
{
public:
	int typeId;
	std::vector<std::shared_ptr<TsObjectField>> fields;

	TsObject(int typeId)
		: TsObject(typeId, std::vector<std::shared_ptr<TsObjectField>>()) {}

	TsObject(int typeId, std::vector<std::shared_ptr<TsObjectField>> fields)
		: typeId(typeId),
		  fields(std::move(fields)) {}

	const std::shared_ptr<TsObjectField> getField(const std::string &field_name) const
	{
		return *std::find_if(fields.begin(), fields.end(), [](auto field)
							 { return *field->descriptor->name == field_name });
	}

	TsObjectFieldDescriptor getFieldDescriptor(const std::string &field_name)
	{
		return getField(field_name)->descriptor;
	}

	std::shared_ptr<TsObject> getFieldValue(const std::string &fieldName)
	{
		return getField(fieldName)->value;
	}

	void setFieldValue(const std::string &field_name, std::shared_ptr<TsObject> value)
	{
		auto field = getField(field_name);

		// If there's no existing field, create one.
		if (field == NULL)
		{
			// Create a new field.
			auto new_field = TsObjectField(TsObjectFieldDescriptor(field_name, value->typeId), value);

			// Attach the field.
			fields.push_back(std::make_shared<TsObjectField>(new_field));

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

	virtual std::shared_ptr<TsObject> invoke(std::vector<TsFunctionArg> args)
	{
		throw std::runtime_error("type is not invocable!");
	}
};

std::shared_ptr<TsObject> ts_main;

int main()
{
	ts_main->invoke(std::vector<TsFunctionArg>());

	return 0;
}
