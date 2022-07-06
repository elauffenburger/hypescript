#include <algorithm>
#include <memory>
#include <stdexcept>
#include <stdio.h>
#include <stdlib.h>
#include <string>
#include <typeinfo>
#include <vector>

#include "runtime.hpp"

// TsNum

TsNum::TsNum(int value) : value(value), TsObject(TypeIdTsNum) {
  this->fields.push_back(new TsObjectField(
      TsObjectFieldDescriptor(TsString("_++"), 0),
      new TsFunction(
          "_++", {},
          [=](TsObject *_this, std::vector<TsFunctionArg> args) -> TsObject * {
            this->value++;

            return _this;
          })));

  this->fields.push_back(new TsObjectField(
      TsObjectFieldDescriptor(TsString("<"), 0),
      new TsFunction(
          "<", {TsFunctionParam("other", 0)},
          [=](TsObject *_this, std::vector<TsFunctionArg> args) -> TsObject * {
            auto other = dynamic_cast<TsNum *>(
                TsFunctionArg::findArg(args, "other").value);

            return new TsBool(this->value < other->value);
          })));
}

// TsBool

TsBool::TsBool(bool value) : value(value), TsObject(TypeIdBool) {}

// TsFunction

TsFunctionParam::TsFunctionParam(std::string name, int type_id)
    : name(name), type_id(type_id) {}

TsFunctionArg::TsFunctionArg(std::string name, TsObject *value)
    : name(name), value(value) {}

const TsFunctionArg &
TsFunctionArg::findArg(const std::vector<TsFunctionArg> &args,
                       const std::string &argName) {
  return *std::find_if(args.begin(), args.end(),
                       [argName](auto arg) { return arg.name == argName; });
}

TsFunction::TsFunction(
    std::string name, std::vector<TsFunctionParam> params,
    std::function<TsObject *(TsObject *, std::vector<TsFunctionArg> args)> fn)
    : name(name), params(params), fn(fn), TsObject(TypeIdTsFunction) {}

TsObjectFieldDescriptor::TsObjectFieldDescriptor(TsString name, int typeId)
    : name(name), typeId(typeId) {}

// TsObject

TsObject::TsObject(int typeId, std::vector<TsObjectField *> fields)
    : typeId(typeId), fields(fields) {
  for (auto field : fields) {
    auto value = field->value;

    // For any functions that don't have an explicit `this` set, use this object
    // as `this`.
    if (value->typeId == TypeIdTsFunction) {
      TsFunction *fn = dynamic_cast<TsFunction *>(value);
      if (fn->thisFn == NULL) {
        fn->thisFn = [this]() -> TsObject * { return this; };
      }
    }
  }
}

TsObjectField::TsObjectField(TsObjectFieldDescriptor descriptor)
    : TsObjectField(descriptor, NULL) {}

TsObjectField::TsObjectField(TsObjectFieldDescriptor descriptor,
                             TsObject *value)
    : descriptor(descriptor), value(value) {}

TsObjectField *TsObject::getField(const std::string &field_name) const {
  return *std::find_if(this->fields.begin(), this->fields.end(),
                       [field_name](TsObjectField *field) {
                         return field->descriptor.name == field_name;
                       });
}

TsObjectFieldDescriptor
TsObject::getFieldDescriptor(const std::string &field_name) const {
  return getField(field_name)->descriptor;
}

TsObject *TsObject::getFieldValue(const std::string &fieldName) const {
  return getField(fieldName)->value;
}

void TsObject::setFieldValue(const std::string &field_name, TsObject *value) {
  auto field = getField(field_name);

  // If there's no existing field, create one.
  if (field == NULL) {
    // Create a new field.
    auto new_field = new TsObjectField(
        TsObjectFieldDescriptor(field_name, value->typeId), value);

    // Attach the field.
    this->fields.push_back(new_field);

    return;
  }

  // Otherwise, just set the field.
  // TODO: handle freeing memory (whoops!).
  field->value = value;
}

template <typename T>
void TsObject::addIntrinsicField(const std::string &fieldName, T value) {
  auto descriptor =
      TsObjectFieldDescriptor(TsString(fieldName), TypeIdIntrinsic);
  auto object = new IntrinsicTsObject<T>(value);
  auto field = new TsObjectField(descriptor, object);

  this->fields.push_back(field);
}

template <typename T>
T TsObject::getIntrinsicField(const std::string &fieldName) const {
  auto field =
      dynamic_cast<IntrinsicTsObject<T> *>(this->getFieldValue(fieldName));

  return field->value;
}

TsObject *TsObject::invoke(std::vector<TsFunctionArg> args) {
  throw std::runtime_error("type is not invocable!");
}

TsObject *TsFunction::invoke(std::vector<TsFunctionArg> args) {
  // If there's no thisFn, use ourself.
  auto _this = this->thisFn == NULL ? this : this->thisFn();

  return fn(_this, args);
}

// Globals

TsObject *console = new TsObject(
    TypeIdTsObject,
    {new TsObjectField(
        TsObjectFieldDescriptor(TsString("log"), TypeIdTsFunction),
        new TsFunction("log", {}, [](auto _this, auto args) -> TsObject * {
          auto fmt = args[0].value->toTsString();

          printf("%s\n", fmt->value.c_str());

          return NULL;
        }))});
