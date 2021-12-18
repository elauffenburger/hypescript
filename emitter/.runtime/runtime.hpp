#pragma once

#include <stdlib.h>
#include <stdio.h>
#include <string>
#include <vector>
#include <algorithm>
#include <memory>

// TODO: don't do this.
#define UNDEFINED 0xDEADBEEF

template <typename T>
class IntrinsicTsObject;

class TsString;
class TsFunctionParam;
class TsFunctionArg;
class TsFunction;
class TsObjectFieldDescriptor;
class TsObjectField;

class TsCoreHelpers
{
public:
    template <class T>
    static std::vector<T> toVector(std::initializer_list<T> args)
    {
        auto result = std::vector<T>();
        for (auto arg : args)
        {
            result.push_back(arg);
        }

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
    TypeIdIntrinsic = 6,
};

class TsObject
{
public:
    int typeId;
    std::vector<TsObjectField *> fields;

    TsObject(int typeId)
        : TsObject(typeId, std::vector<TsObjectField *>()) {}

    TsObject(int typeId, std::vector<TsObjectField *> fields)
        : typeId(typeId),
          fields(fields) {}

    TsObjectField *getField(const std::string &field_name) const;

    TsObjectFieldDescriptor getFieldDescriptor(const std::string &field_name) const;

    TsObject *getFieldValue(const std::string &fieldName) const;

    void setFieldValue(const std::string &field_name, TsObject *value);

    template <typename T>
    void addIntrinsicField(const std::string &fieldName, T value);

    template <typename T>
    T getIntrinsicField(const std::string &fieldName) const;

    virtual TsObject *invoke(std::vector<TsFunctionArg> args);
};

template <typename T>
class IntrinsicTsObject : public TsObject
{
public:
    T value;

    IntrinsicTsObject(T value)
        : value(value),
          TsObject(TypeIdIntrinsic) {}

    bool operator==(const IntrinsicTsObject<T> &other) const
    {
        return value == other.value;
    }
};

class TsNum : public TsObject
{
public:
    int num;

    TsNum(int num)
        : num(num),
          TsObject(TypeIdTsNum) {}

    bool operator==(const TsNum &other) const
    {
        return num == other.num;
    }
};

class TsString : public TsObject
{
public:
    std::string value;

    TsString(std::string value)
        : value(value),
          TsObject(TypeIdTsString) {}

    bool operator==(const TsString &other) const
    {
        return value == other.value;
    }

    bool operator==(const std::string &other) const
    {
        return value == other;
    }
};

class TsFunctionParam
{
public:
    std::string name;
    int type_id;

    TsFunctionParam(std::string name, int type_id)
        : name(name),
          type_id(type_id) {}
};

class TsFunctionArg
{
public:
    std::string name;
    TsObject *value;

    TsFunctionArg(std::string name, TsObject *value)
        : name(name),
          value(value) {}

    static const TsFunctionArg &findArg(const std::vector<TsFunctionArg> &args, const std::string &argName);
};

typedef TsObject *(*TsFunctionFn)(std::vector<TsFunctionArg> args);

class TsFunction : public TsObject
{
public:
    std::string name;
    std::vector<TsFunctionParam> params;

    TsFunctionFn fn;

    TsFunction(std::string name, std::vector<TsFunctionParam> params, TsFunctionFn fn)
        : name(name),
          params(params),
          fn(fn),
          TsObject(TypeIdTsFunction) {}

    virtual TsObject *invoke(std::vector<TsFunctionArg> args) override;
};

class TsObjectFieldDescriptor
{
public:
    TsString name;
    int typeId;

    TsObjectFieldDescriptor(TsString name, int typeId)
        : name(name),
          typeId(typeId) {}
};

class TsObjectField
{
public:
    TsObjectFieldDescriptor descriptor;
    TsObject *value;

    TsObjectField(TsObjectFieldDescriptor descriptor)
        : TsObjectField(descriptor, NULL) {}

    TsObjectField(TsObjectFieldDescriptor descriptor, TsObject *value)
        : descriptor(descriptor),
          value(value) {}
};

extern TsFunction *ts_fn_main;
