#include <stdlib.h>
#include <stdio.h>
#include <string>
#include <vector>
#include <algorithm>
#include <memory>
#include <stdexcept>

#include "runtime.hpp"

TsObjectField *TsObject::getField(const std::string &field_name) const
{
	return *std::find_if(this->fields.begin(), this->fields.end(), [field_name](TsObjectField *field)
						 { return field->descriptor.name == field_name; });
}

TsObjectFieldDescriptor TsObject::getFieldDescriptor(const std::string &field_name) const
{
	return getField(field_name)->descriptor;
}

TsObject *TsObject::getFieldValue(const std::string &fieldName) const
{
	return getField(fieldName)->value;
}

void TsObject::setFieldValue(const std::string &field_name, TsObject *value)
{
	auto field = getField(field_name);

	// If there's no existing field, create one.
	if (field == NULL)
	{
		// Create a new field.
		auto new_field = new TsObjectField(TsObjectFieldDescriptor(field_name, value->typeId), value);

		// Attach the field.
		this->fields.push_back(new_field);

		return;
	}

	// Otherwise, just set the field.
	// TODO: handle freeing memory (whoops!).
	field->value = value;
}

template <typename T>
void TsObject::addIntrinsicField(const std::string &fieldName, T value)
{
	auto descriptor = TsObjectFieldDescriptor(TsString(fieldName), TypeIdIntrinsic);
	auto object = new IntrinsicTsObject<T>(value);
	auto field = new TsObjectField(descriptor, object);

	this->fields.push_back(field);
}

template <typename T>
T TsObject::getIntrinsicField(const std::string &fieldName) const
{
	auto field = dynamic_cast<IntrinsicTsObject<T> *>(this->getFieldValue(fieldName));

	return field->value;
}

TsObject *TsObject::invoke(std::vector<TsFunctionArg> args)
{
	throw std::runtime_error("type is not invocable!");
}

const TsFunctionArg &TsFunctionArg::findArg(const std::vector<TsFunctionArg> &args, const std::string &argName)
{
	return *std::find_if(args.begin(), args.end(), [argName](auto arg)
						 { return arg.name == argName; });
}

TsObject *TsFunction::invoke(std::vector<TsFunctionArg> args)
{
	return fn(args);
}

TsObject* console = new TsObject(TypeIdTsObject, TsCoreHelpers::toVector<TsObjectField *>({new TsObjectField(
												TsObjectFieldDescriptor(TsString("log"), TypeIdTsFunction),
												new TsFunction("log",
															   TsCoreHelpers::toVector<TsFunctionParam>({}),
															   [](std::vector<TsFunctionArg> args) -> TsObject *
															   {
																   auto fmt = dynamic_cast<TsString *>(args[0].value);

																   printf("%s\n", fmt->value.c_str());

																   return NULL;
															   }))}));
