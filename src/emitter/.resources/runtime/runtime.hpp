#pragma once

#include <algorithm>
#include <functional>
#include <memory>
#include <stdio.h>
#include <stdlib.h>
#include <string>
#include <vector>

// TODO: don't do this.
#define UNDEFINED 0xDEADBEEF

template <typename T> class IntrinsicTsObject;

class TsString;
class TsFunctionParam;
class TsFunctionArg;
class TsFunction;
class TsObjectFieldDescriptor;
class TsObjectField;

enum TypeId {
  TypeIdNone = 0,
  TypeIdTsObject = 1,
  TypeIdTsNum = 2,
  TypeIdTsString = 3,
  TypeIdTsFunction = 4,
  TypeIdVoid = 5,
  TypeIdIntrinsic = 6,
  TypeIdBool = 7,
};

class TsObject {
public:
  int typeId;
  std::vector<TsObjectField *> fields;

  TsObject(int typeId) : TsObject(typeId, std::vector<TsObjectField *>()) {}

  TsObject(int typeId, std::vector<TsObjectField *> fields);

  TsObjectField *getField(const std::string &field_name) const;

  TsObjectFieldDescriptor
  getFieldDescriptor(const std::string &field_name) const;

  TsObject *getFieldValue(const std::string &fieldName) const;

  void setFieldValue(const std::string &field_name, TsObject *value);

  template <typename T>
  void addIntrinsicField(const std::string &fieldName, T value);

  template <typename T> T getIntrinsicField(const std::string &fieldName) const;

  virtual TsObject *invoke(std::vector<TsFunctionArg> args);

  virtual bool truthy() { throw "not implemented!"; }

  virtual TsString *toTsString() { return NULL; }
};

class TsString : public TsObject {
public:
  std::string value;

  TsString(std::string value) : value(value), TsObject(TypeIdTsString) {}

  bool operator==(const TsString &other) const { return value == other.value; }

  bool operator==(const std::string &other) const { return value == other; }

  virtual bool truthy() override { return value != ""; }

  virtual TsString *toTsString() override { return this; }
};

template <typename T> class IntrinsicTsObject : public TsObject {
public:
  T value;

  IntrinsicTsObject(T value) : value(value), TsObject(TypeIdIntrinsic) {}

  bool operator==(const IntrinsicTsObject<T> &other) const {
    return value == other.value;
  }

  virtual TsString *toTsString() override { return new TsString("object"); }
};

class TsNum : public TsObject {
public:
  float value;

  TsNum(float value);

  bool operator==(const TsNum &other) const { return value == other.value; }

  virtual TsString *toTsString() override {
    if (value == (int)value) {
      return new TsString(std::to_string((int)value));
    }

    return new TsString(std::to_string(value));
  }

  virtual bool truthy() override { return value != 0; }
};

class TsBool : public TsObject {
public:
  bool value;

  TsBool(bool value);

  bool operator==(const TsBool &other) const { return value == other.value; }

  operator bool() const { return value; }

  virtual TsString *toTsString() override {
    return new TsString(value ? "true" : "false");
  }

  virtual bool truthy() override { return value; }
};

class TsFunctionParam {
public:
  std::string name;
  int type_id;

  TsFunctionParam(std::string name, int type_id);
};

class TsFunctionArg {
public:
  std::string name;
  TsObject *value;

  TsFunctionArg(std::string name, TsObject *value);

  static const TsFunctionArg &findArg(const std::vector<TsFunctionArg> &args,
                                      const std::string &argName);
};

class TsFunction : public TsObject {
public:
  std::string name;
  std::vector<TsFunctionParam> params;
  std::function<TsObject *()> thisFn;

  std::function<TsObject *(TsObject *, std::vector<TsFunctionArg> args)> fn;

  TsFunction(
      std::string name, std::vector<TsFunctionParam> params,
      std::function<TsObject *(TsObject *, std::vector<TsFunctionArg> args)>
          fn);

  virtual TsObject *invoke(std::vector<TsFunctionArg> args) override;

  virtual TsString *toTsString() override {
    return new TsString("function<" + name + ">");
  }

  virtual bool truthy() override { return true; }
};

class TsObjectFieldDescriptor {
public:
  std::string name;
  int typeId;

  TsObjectFieldDescriptor(std::string name, int typeId);
};

class TsObjectField {
public:
  TsObjectFieldDescriptor descriptor;
  TsObject *value;

  TsObjectField(TsObjectFieldDescriptor descriptor);

  TsObjectField(TsObjectFieldDescriptor descriptor, TsObject *value);
};

extern TsObject *console;
