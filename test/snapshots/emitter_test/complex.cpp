
		#include <stdlib.h>
		#include <stdio.h>
		#include <string>
		#include <vector>
		#include <algorithm>
		#include <memory>
	
		#include "runtime.hpp"

		int main()
		{
			ts_fn_main->invoke(std::vector<TsFunctionArg>());

			return 0;
		}
	auto ts_fn_foo = new TsFunction("foo", TsCoreHelpers::toVector<TsFunctionParam>({TsFunctionParam("a", 0), TsFunctionParam("b", 0)}), [](std::vector<TsFunctionArg> args) -> TsObject* {auto a = TsFunctionArg::findArg(args, "a").value;auto b = TsFunctionArg::findArg(args, "b").value;auto ay = new TsNum(5);auto bee = new TsString("bar");return ay;});


auto ts_fn_blah = new TsFunction("blah", TsCoreHelpers::toVector<TsFunctionParam>({}), [](std::vector<TsFunctionArg> args) -> TsObject* {auto foo = new TsString("asdf");auto bar = foo;bar = new TsString("bar");return foo;});


auto ts_fn_blah2 = new TsFunction("blah2", TsCoreHelpers::toVector<TsFunctionParam>({}), [](std::vector<TsFunctionArg> args) -> TsObject* {return NULL;});


auto ts_fn_main = new TsFunction("main", TsCoreHelpers::toVector<TsFunctionParam>({}), [](std::vector<TsFunctionArg> args) -> TsObject* {auto obj = new TsObject(2, TsCoreHelpers::toVector<TsObjectField*>({new TsObjectField(TsObjectFieldDescriptor(TsString("foo"), 0), new TsString("bar")), new TsObjectField(TsObjectFieldDescriptor(TsString("baz"), 0), new TsNum(5)), new TsObjectField(TsObjectFieldDescriptor(TsString("qux"), 0), new TsObject(2, TsCoreHelpers::toVector<TsObjectField*>({new TsObjectField(TsObjectFieldDescriptor(TsString("a"), 0), new TsString("a"))})))}));obj->getFieldValue("qux")->setFieldValue("a", new TsString("b"));ts_fn_blah->invoke(TsCoreHelpers::toVector<TsFunctionArg>({}));return NULL;});


