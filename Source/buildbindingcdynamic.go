/*++

Copyright (C) 2018 Autodesk Inc. (Original Author)

All rights reserved.

Redistribution and use in source and binary forms, with or without modification,
are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this
list of conditions and the following disclaimer.
2. Redistributions in binary form must reproduce the above copyright notice,
this list of conditions and the following disclaimer in the documentation
and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

--*/

//////////////////////////////////////////////////////////////////////////////////////////////////////
// buildbindingcdynamic.go
// functions to generate dynamic C-bindings of a library's API in form of dynamically loaded functions
// handles.
//////////////////////////////////////////////////////////////////////////////////////////////////////

package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
)

// BuildBindingCDynamic builds dyanmic C-bindings of a library's API in form of dynamically loaded functions
// handles.
func BuildBindingCDynamic(component ComponentDefinition, outputFolder string) error {

	namespace := component.NameSpace;
	libraryname := component.LibraryName;
	baseName := component.BaseName;

	DynamicCHeader := path.Join(outputFolder, baseName+"_dynamic.h");
	log.Printf("Creating \"%s\"", DynamicCHeader)
	dynhfile, err := os.Create(DynamicCHeader)
	if err != nil {
		return err;
	}
	WriteLicenseHeader(dynhfile, component,
		fmt.Sprintf("This is an autogenerated plain C Header file in order to allow an easy\n use of %s", libraryname),
		true)
	err = buildDynamicCHeader(component, dynhfile, namespace, baseName, false)
	if err != nil {
		return err;
	}

	DynamicCImpl := path.Join(outputFolder, baseName+"_dynamic.cpp");
	log.Printf("Creating \"%s\"", DynamicCImpl)
	dyncppfile, err := os.Create(DynamicCImpl)
	if err != nil {
		return err;
	}
	WriteLicenseHeader(dyncppfile, component,
		fmt.Sprintf("This is an autogenerated plain C Header file in order to allow an easy\n use of %s", libraryname),
		true)
	
	err = buildDynamicCppImplementation(component, dyncppfile, namespace, baseName)
	if err != nil {
		return err;
	}
	
	return nil;
}

func buildDynamicCHeader(component ComponentDefinition, w io.Writer, NameSpace string, BaseName string, headerOnly bool) error {
	fmt.Fprintf(w, "#ifndef __%s_DYNAMICHEADER\n", strings.ToUpper(NameSpace))
	fmt.Fprintf(w, "#define __%s_DYNAMICHEADER\n", strings.ToUpper(NameSpace))
	fmt.Fprintf(w, "\n")

	fmt.Fprintf(w, "#include \"%s_types.h\"\n", BaseName)
	fmt.Fprintf(w, "\n")

	for i := 0; i < len(component.Classes); i++ {
		class := component.Classes[i]

		fmt.Fprintf(w, "\n")
		fmt.Fprintf(w, "/*************************************************************************************************************************\n")
		fmt.Fprintf(w, " Class definition for %s\n", class.ClassName)
		fmt.Fprintf(w, "**************************************************************************************************************************/\n")

		for j := 0; j < len(class.Methods); j++ {
			method := class.Methods[j]
			WriteCMethod(method, w, NameSpace, class.ClassName, false, true)
		}

	}

	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "/*************************************************************************************************************************\n")
	fmt.Fprintf(w, " Global functions\n")
	fmt.Fprintf(w, "**************************************************************************************************************************/\n")

	global := component.Global;
	for j := 0; j < len(global.Methods); j++ {
		method := global.Methods[j]
		err := WriteCMethod(method, w, NameSpace, "Wrapper", true, true)
		if err != nil {
			return err
		}

	}

	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "/*************************************************************************************************************************\n")
	fmt.Fprintf(w, " Function Table Structure\n")
	fmt.Fprintf(w, "**************************************************************************************************************************/\n")
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "typedef struct {\n")
	fmt.Fprintf(w, "    void * m_LibraryHandle;\n")

	for i := 0; i < len(component.Classes); i++ {
		class := component.Classes[i]
		for j := 0; j < len(class.Methods); j++ {
			method := class.Methods[j]
			fmt.Fprintf(w, "    P%s%s_%sPtr m_%s_%s;\n", NameSpace, class.ClassName, method.MethodName, class.ClassName, method.MethodName)
		}
	}

	for j := 0; j < len(global.Methods); j++ {
		method := global.Methods[j]
		fmt.Fprintf(w, "    P%s%sPtr m_%s;\n", NameSpace, method.MethodName, method.MethodName)
	}

	fmt.Fprintf(w, "} s%sDynamicWrapperTable;\n", NameSpace)
	fmt.Fprintf(w, "\n")
	
	if (!headerOnly) {

		fmt.Fprintf(w, "/*************************************************************************************************************************\n")
		fmt.Fprintf(w, " Load DLL dynamically\n")
		fmt.Fprintf(w, "**************************************************************************************************************************/\n")

		fmt.Fprintf(w, "%sResult Init%sWrapperTable (s%sDynamicWrapperTable * pWrapperTable);\n", NameSpace, NameSpace, NameSpace)
		fmt.Fprintf(w, "%sResult Release%sWrapperTable (s%sDynamicWrapperTable * pWrapperTable);\n", NameSpace, NameSpace, NameSpace)
		fmt.Fprintf(w, "%sResult Load%sWrapperTable (s%sDynamicWrapperTable * pWrapperTable, const char * pLibraryFileName);\n", NameSpace, NameSpace, NameSpace)

		fmt.Fprintf(w, "\n")
	}
	
	fmt.Fprintf(w, "#endif // __%s_DYNAMICHEADER\n", strings.ToUpper(NameSpace))
	fmt.Fprintf(w, "\n")

	return nil
}


func buildDynamicCInitTableCode(component ComponentDefinition, w io.Writer, NameSpace string, BaseName string, spacing string) error {
	global := component.Global;
	
	fmt.Fprintf(w, "%sif (pWrapperTable == nullptr)\n", spacing)
	fmt.Fprintf(w, "%s    return %s_ERROR_INVALIDPARAM;\n", spacing, strings.ToUpper(NameSpace))
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "%spWrapperTable->m_LibraryHandle = nullptr;\n", spacing)

	for i := 0; i < len(component.Classes); i++ {
		class := component.Classes[i]
		for j := 0; j < len(class.Methods); j++ {
			method := class.Methods[j]
			fmt.Fprintf(w, "%spWrapperTable->m_%s_%s = nullptr;\n", spacing, class.ClassName, method.MethodName)
		}
	}

	global = component.Global
	for j := 0; j < len(global.Methods); j++ {
		method := global.Methods[j]
		fmt.Fprintf(w, "%spWrapperTable->m_%s = nullptr;\n", spacing, method.MethodName)
	}

	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "%sreturn %s_SUCCESS;\n", spacing, strings.ToUpper(NameSpace))
	
	return nil;
}


func buildDynamicCReleaseTableCode(component ComponentDefinition, w io.Writer, NameSpace string, BaseName string, spacing string, initWrapperFunctionName string) error {

	fmt.Fprintf(w, "%sif (pWrapperTable == nullptr)\n", spacing)
	fmt.Fprintf(w, "%s    return %s_ERROR_INVALIDPARAM;\n", spacing, strings.ToUpper(NameSpace))
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "%sif (pWrapperTable->m_LibraryHandle != nullptr) {\n", spacing)
	fmt.Fprintf(w, "%s    HMODULE hModule = (HMODULE) pWrapperTable->m_LibraryHandle;\n", spacing)
	fmt.Fprintf(w, "%s    FreeLibrary (hModule);\n", spacing)
	fmt.Fprintf(w, "%s    return %s (pWrapperTable);\n", spacing, initWrapperFunctionName)
	fmt.Fprintf(w, "%s}\n", spacing)
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "%sreturn %s_SUCCESS;\n", spacing, strings.ToUpper(NameSpace))

	return nil;
}


func buildDynamicCLoadTableCode(component ComponentDefinition, w io.Writer, NameSpace string, BaseName string, spacing string) error {

	global := component.Global;

	fmt.Fprintf(w, "%sif (pWrapperTable == nullptr)\n", spacing)
	fmt.Fprintf(w, "%s    return %s_ERROR_INVALIDPARAM;\n", spacing, strings.ToUpper(NameSpace))
	fmt.Fprintf(w, "%sif (pLibraryFileName == nullptr)\n", spacing)
	fmt.Fprintf(w, "%s    return %s_ERROR_INVALIDPARAM;\n", spacing, strings.ToUpper(NameSpace))

	fmt.Fprintf(w, "\n")

	// TODO: Unicode
	fmt.Fprintf(w, "%sHMODULE hLibrary = LoadLibraryA (pLibraryFileName);\n", spacing)
	fmt.Fprintf(w, "%sif (hLibrary == 0) \n", spacing)
	fmt.Fprintf(w, "%s    return %s_ERROR_COULDNOTLOADLIBRARY;\n", spacing, strings.ToUpper(NameSpace))

	for i := 0; i < len(component.Classes); i++ {

		class := component.Classes[i]
		for j := 0; j < len(class.Methods); j++ {

			method := class.Methods[j]

			fmt.Fprintf(w, "%spWrapperTable->m_%s_%s = (P%s%s_%sPtr) GetProcAddress (hLibrary, \"%s_%s_%s%s\");\n", spacing, class.ClassName, method.MethodName, NameSpace, class.ClassName, method.MethodName, strings.ToLower(NameSpace), strings.ToLower(class.ClassName), strings.ToLower(method.MethodName), method.DLLSuffix)
			fmt.Fprintf(w, "%sif (pWrapperTable->m_%s_%s == nullptr)\n", spacing, class.ClassName, method.MethodName)
			fmt.Fprintf(w, "%s    return %s_ERROR_COULDNOTFINDLIBRARYEXPORT;\n", spacing, strings.ToUpper(NameSpace))
			fmt.Fprintf(w, "\n")
		}
	}

	global = component.Global
	for j := 0; j < len(global.Methods); j++ {
		method := global.Methods[j]

		fmt.Fprintf(w, "%spWrapperTable->m_%s = (P%s%sPtr) GetProcAddress (hLibrary, \"%s_%s%s\");\n", spacing, method.MethodName, NameSpace, method.MethodName, strings.ToLower(NameSpace), strings.ToLower(method.MethodName), method.DLLSuffix)
		fmt.Fprintf(w, "%sif (pWrapperTable->m_%s == nullptr)\n", spacing, method.MethodName)
		fmt.Fprintf(w, "%s    return %s_ERROR_COULDNOTFINDLIBRARYEXPORT;\n", spacing, strings.ToUpper(NameSpace))
		fmt.Fprintf(w, "\n")
	}

	fmt.Fprintf(w, "%spWrapperTable->m_LibraryHandle = hLibrary;\n", spacing)
	fmt.Fprintf(w, "%sreturn %s_SUCCESS;\n", spacing, strings.ToUpper(NameSpace))

	return nil;
}

func buildDynamicCppImplementation(component ComponentDefinition, w io.Writer, NameSpace string, BaseName string) error {

	fmt.Fprintf(w, "#include \"%s_types.h\"\n", BaseName)
	fmt.Fprintf(w, "#include \"%s_dynamic.h\"\n", BaseName)
	fmt.Fprintf(w, "#include <Windows.h>\n")

	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "%sResult Init%sWrapperTable (s%sDynamicWrapperTable * pWrapperTable)\n", NameSpace, NameSpace, NameSpace)
	fmt.Fprintf(w, "{\n")
	
	buildDynamicCInitTableCode (component, w, NameSpace, BaseName, "    ");
	
	fmt.Fprintf(w, "}\n")

	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "%sResult Release%sWrapperTable (s%sDynamicWrapperTable * pWrapperTable)\n", NameSpace, NameSpace, NameSpace)
	fmt.Fprintf(w, "{\n")
	
	buildDynamicCReleaseTableCode (component, w, NameSpace, BaseName, "    ", "Init" + NameSpace + "WrapperTable");
	
	fmt.Fprintf(w, "}\n")

	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "%sResult Load%sWrapperTable (s%sDynamicWrapperTable * pWrapperTable, const char * pLibraryFileName)\n", NameSpace, NameSpace, NameSpace)
	fmt.Fprintf(w, "{\n")
	
	buildDynamicCLoadTableCode (component, w, NameSpace, BaseName, "    ");
	
	fmt.Fprintf(w, "}\n")
	fmt.Fprintf(w, "\n")

	return nil
}


func writeDynamicCPPMethodDefinition(method ComponentDefinitionMethod, w io.Writer, NameSpace string, ClassName string, isGlobal bool) error {
	parameters := ""
	returntype := "void"

	for k := 0; k < len(method.Params); k++ {

		param := method.Params[k]
		variableName := getBindingCppVariableName(param)

		switch param.ParamPass {
		case "in":

			if parameters != "" {
				parameters = parameters + ", "
			}

			cppParamType := getBindingCppParamType(param, NameSpace, true)

			switch param.ParamType {
			case "string":
				parameters = parameters + fmt.Sprintf("const %s & %s", cppParamType, variableName);
			case "struct":
				parameters = parameters + fmt.Sprintf("const %s & %s", cppParamType, variableName);
			case "structarray", "basicarray":
				parameters = parameters + fmt.Sprintf("const %s & %s", cppParamType, variableName);
			case "handle":
				parameters = parameters + fmt.Sprintf("%s %s", cppParamType, variableName)

			default:
				parameters = parameters + fmt.Sprintf("const %s %s", cppParamType, variableName)
			}

		case "out":
			cppParamType := getBindingCppParamType(param, NameSpace, false)
	
			if parameters != "" {
				parameters = parameters + ", "
			}
			parameters = parameters + fmt.Sprintf("%s & %s", cppParamType, variableName)


		case "return":
			returntype = getBindingCppParamType(param, NameSpace, false)

		default:
			return fmt.Errorf("invalid method parameter passing \"%s\" for %s.%s (%s)", param.ParamPass, ClassName, method.MethodName, param.ParamName)
		}

	}
	
	fmt.Fprintf(w, "    %s %s (%s);\n", returntype, method.MethodName, parameters);

	return nil
}



func writeDynamicCPPMethod(method ComponentDefinitionMethod, w io.Writer, NameSpace string, ClassName string, isGlobal bool) error {

	CMethodName := ""
	requiresInitCall := false;
	initCallParameters := ""	// usually used to check sizes of buffers
	callParameters := ""
	checkErrorCode := ""
	makeSharedParameter := "";

	if isGlobal {
		CMethodName = fmt.Sprintf("m_WrapperTable.m_%s", method.MethodName)
		checkErrorCode = "CheckError (nullptr,"
		makeSharedParameter = "this, "
	} else {
		CMethodName = fmt.Sprintf("m_pWrapper->m_WrapperTable.m_%s_%s", ClassName, method.MethodName)
		callParameters = "m_pHandle"
		initCallParameters = "m_pHandle"
		checkErrorCode = "CheckError ("
		makeSharedParameter = "m_pWrapper, "
	}

	parameters := ""
	returntype := "void"

	definitionCode := "";
	functioncode := ""
	returncode := ""
	commentcode := ""
	postCallCode := ""

	cppClassPrefix := "C" + NameSpace
	cppClassName := cppClassPrefix + ClassName

	for k := 0; k < len(method.Params); k++ {

		param := method.Params[k]
		variableName := getBindingCppVariableName(param)

		callParameter := "";
		initCallParameter := "";

		switch param.ParamPass {
		case "in":

			if parameters != "" {
				parameters = parameters + ", "
			}

			cppParamType := getBindingCppParamType(param, NameSpace, true)
			commentcode = commentcode + fmt.Sprintf("    * @param[in] %s - %s\n", variableName, param.ParamDescription)

			switch param.ParamType {
			case "string":
				callParameter = variableName + ".c_str()"
				initCallParameter = callParameter;
				parameters = parameters + fmt.Sprintf("const %s & %s", cppParamType, variableName);
			case "struct":
				callParameter = "&" + variableName
				initCallParameter = callParameter;
				parameters = parameters + fmt.Sprintf("const %s & %s", cppParamType, variableName);
			case "structarray", "basicarray":
				callParameter = fmt.Sprintf("(unsigned int)%s.size(), %s.data()", variableName, variableName);
				initCallParameter = callParameter;
				parameters = parameters + fmt.Sprintf("const %s & %s", cppParamType, variableName);
			case "handle":
				functioncode = functioncode + fmt.Sprintf("        %sHandle h%s = nullptr;\n", NameSpace, param.ParamName)
				functioncode = functioncode + fmt.Sprintf("        if (%s != nullptr) {\n", variableName)
				functioncode = functioncode + fmt.Sprintf("            h%s = %s->GetHandle ();\n", param.ParamName, variableName)
				functioncode = functioncode + fmt.Sprintf("        };\n")
				callParameter = "h" + param.ParamName;
				initCallParameter = callParameter;
				parameters = parameters + fmt.Sprintf("%s %s", cppParamType, variableName)

			default:
				callParameter = variableName;
				initCallParameter = callParameter;
				parameters = parameters + fmt.Sprintf("const %s %s", cppParamType, variableName)
			}

		case "out":
			cppParamType := getBindingCppParamType(param, NameSpace, false)
			commentcode = commentcode + fmt.Sprintf("    * @param[out] %s - %s\n", variableName, param.ParamDescription)

			if parameters != "" {
				parameters = parameters + ", "
			}
			parameters = parameters + fmt.Sprintf("%s & %s", cppParamType, variableName)

			switch param.ParamType {

			case "string":
				requiresInitCall = true;
				definitionCode = definitionCode + fmt.Sprintf("        unsigned int bytesNeeded%s = 0;\n", param.ParamName)
				definitionCode = definitionCode + fmt.Sprintf("        unsigned int bytesWritten%s = 0;\n", param.ParamName)
				initCallParameter = fmt.Sprintf("0, &bytesNeeded%s, nullptr", param.ParamName);
				
				functioncode = functioncode + fmt.Sprintf("        std::vector<char> buffer%s;\n", param.ParamName)
				functioncode = functioncode + fmt.Sprintf("        buffer%s.resize(bytesNeeded%s + 2);\n", param.ParamName, param.ParamName)

				callParameter = fmt.Sprintf("bytesNeeded%s + 2, &bytesWritten%s, &buffer%s[0]", param.ParamName, param.ParamName, param.ParamName)

				postCallCode = postCallCode + fmt.Sprintf("        buffer%s[bytesNeeded%s + 1] = 0;\n", param.ParamName, param.ParamName) +
					fmt.Sprintf("        s%s = std::string(&buffer%s[0]);\n", param.ParamName, param.ParamName)

			case "handle":
				// NOTTESTED
				definitionCode = definitionCode + fmt.Sprintf("        %sHandle h%s = nullptr;\n", NameSpace, param.ParamName)
				callParameter = fmt.Sprintf("&h%s", param.ParamName)
				initCallParameter = callParameter;
				postCallCode = postCallCode + fmt.Sprintf("        p%s = std::make_shared<%s%s> (h%s);\n", param.ParamName, cppClassPrefix, param.ParamClass, param.ParamName)

			case "structarray", "basicarray":
				requiresInitCall = true;
				definitionCode = definitionCode + fmt.Sprintf("        unsigned int elementsNeeded%s = 0;\n", param.ParamName)
				definitionCode = definitionCode + fmt.Sprintf("        unsigned int elementsWritten%s = 0;\n", param.ParamName)
				initCallParameter = fmt.Sprintf("0, &elementsNeeded%s, nullptr", param.ParamName);

				functioncode = functioncode + fmt.Sprintf("        %s.resize(elementsNeeded%s);\n", variableName, param.ParamName);
				callParameter = fmt.Sprintf("elementsNeeded%s, &elementsWritten%s, %s.data()", param.ParamName, param.ParamName, variableName)

			default:
				callParameter = "&" + variableName
				initCallParameter = callParameter
			}

		case "return":

			commentcode = commentcode + fmt.Sprintf("    * @return %s\n", param.ParamDescription)
			returntype = getBindingCppParamType(param, NameSpace, false)

			switch param.ParamType {
			case "uint8", "uint16", "uint32", "uint64", "int8", "int16", "int32", "int64", "bool", "single", "double":
				callParameter = fmt.Sprintf("&result%s", param.ParamName)
				initCallParameter = callParameter;
				definitionCode = definitionCode + fmt.Sprintf("        %s result%s = 0;\n", returntype, param.ParamName)
				returncode = fmt.Sprintf("        return result%s;\n", param.ParamName)

			case "string":
				requiresInitCall = true;
				definitionCode = definitionCode + fmt.Sprintf("        unsigned int bytesNeeded%s = 0;\n", param.ParamName)
				definitionCode = definitionCode + fmt.Sprintf("        unsigned int bytesWritten%s = 0;\n", param.ParamName)
				initCallParameter = fmt.Sprintf("0, &bytesNeeded%s, nullptr", param.ParamName);

				functioncode = functioncode + fmt.Sprintf("        std::vector<char> buffer%s;\n", param.ParamName)
				functioncode = functioncode + fmt.Sprintf("        buffer%s.resize(bytesNeeded%s + 2);\n", param.ParamName, param.ParamName)

				callParameter = fmt.Sprintf("bytesNeeded%s + 2, &bytesWritten%s, &buffer%s[0]", param.ParamName, param.ParamName, param.ParamName)

				returncode = fmt.Sprintf("        buffer%s[bytesNeeded%s + 1] = 0;\n", param.ParamName, param.ParamName) +
					fmt.Sprintf("        return std::string(&buffer%s[0]);\n", param.ParamName)

			case "enum":
				callParameter = fmt.Sprintf("&result%s", param.ParamName)
				initCallParameter = callParameter;
				definitionCode = definitionCode + fmt.Sprintf("        e%s%s result%s = (e%s%s) 0;\n", NameSpace, param.ParamClass, param.ParamName, NameSpace, param.ParamClass)
				returncode = fmt.Sprintf("        return result%s;\n", param.ParamName)

			case "struct":
				callParameter = fmt.Sprintf("&result%s", param.ParamName)
				initCallParameter = callParameter;
				definitionCode = definitionCode + fmt.Sprintf("        s%s%s result%s;\n", NameSpace, param.ParamClass, param.ParamName)
				returncode = fmt.Sprintf("        return result%s;\n", param.ParamName)

			case "handle":
				definitionCode = definitionCode + fmt.Sprintf("        %sHandle h%s = nullptr;\n", NameSpace, param.ParamName)
				callParameter = fmt.Sprintf("&h%s", param.ParamName)
				initCallParameter = callParameter;
				returncode = fmt.Sprintf("        return std::make_shared<%s%s> (%sh%s);\n", cppClassPrefix, param.ParamClass, makeSharedParameter, param.ParamName)

			case "basicarray":
				return fmt.Errorf("can not return basicarray \"%s\" for %s.%s (%s)", param.ParamPass, ClassName, method.MethodName, param.ParamName)

			case "structarray":
				return fmt.Errorf("can not return structarray \"%s\" for %s.%s (%s)", param.ParamPass, ClassName, method.MethodName, param.ParamName)

			default:
				return fmt.Errorf("invalid method parameter type \"%s\" for %s.%s (%s)", param.ParamType, ClassName, method.MethodName, param.ParamName)
			}

		default:
			return fmt.Errorf("invalid method parameter passing \"%s\" for %s.%s (%s)", param.ParamPass, ClassName, method.MethodName, param.ParamName)
		}

		if callParameters != "" {
			callParameters = callParameters + ", "
		}
		callParameters = callParameters + callParameter;
		if (initCallParameters != "") {
			initCallParameters = initCallParameters + ", ";
		}
		initCallParameters = initCallParameters + initCallParameter;

	}

	fmt.Fprintf(w, "    \n")
	fmt.Fprintf(w, "    /**\n")
	fmt.Fprintf(w, "    * %s::%s - %s\n", cppClassName, method.MethodName, method.MethodDescription)
	fmt.Fprintf(w, commentcode)
	fmt.Fprintf(w, "    */\n")
	
	if (isGlobal) {	
		fmt.Fprintf(w, "    inline %s %s::%s (%s)\n", returntype, cppClassName, method.MethodName, parameters)
	} else {
		fmt.Fprintf(w, "    %s %s (%s)\n", returntype, method.MethodName, parameters)
	}

	fmt.Fprintf(w, "    {\n")
	fmt.Fprintf(w, definitionCode)
	if (requiresInitCall) {
		fmt.Fprintf(w, "        %s %s (%s) );\n", checkErrorCode, CMethodName, initCallParameters)
	}
	fmt.Fprintf(w, functioncode)
	fmt.Fprintf(w, "        %s %s (%s) );\n", checkErrorCode, CMethodName, callParameters)
	fmt.Fprintf(w, postCallCode)
	fmt.Fprintf(w, returncode)
	fmt.Fprintf(w, "    }\n")

	return nil
}



func buildDynamicCppHeader(component ComponentDefinition, w io.Writer, NameSpace string, BaseName string) error {

	global := component.Global
	
	cppClassPrefix := "C" + NameSpace
	
	fmt.Fprintf(w, "#ifndef __%s_DYNAMICCPPHEADER\n", strings.ToUpper(NameSpace))
	fmt.Fprintf(w, "#define __%s_DYNAMICCPPHEADER\n", strings.ToUpper(NameSpace))
	fmt.Fprintf(w, "\n")

	fmt.Fprintf(w, "#include \"%s_types.h\"\n", BaseName)
	fmt.Fprintf(w, "#include \"%s_dynamic.h\"\n", BaseName)
	fmt.Fprintf(w, "\n")

	fmt.Fprintf(w, "#include <Windows.h>\n")
	fmt.Fprintf(w, "#include <string>\n")
	fmt.Fprintf(w, "#include <memory>\n")
	fmt.Fprintf(w, "#include <vector>\n")
	fmt.Fprintf(w, "#include <exception>\n")
	fmt.Fprintf(w, "\n")

	fmt.Fprintf(w, "namespace %s {\n", NameSpace)
	fmt.Fprintf(w, "\n")

	fmt.Fprintf(w, "/*************************************************************************************************************************\n")
	fmt.Fprintf(w, " Forward Declaration of all classes \n")
	fmt.Fprintf(w, "**************************************************************************************************************************/\n")
	fmt.Fprintf(w, "\n")

	fmt.Fprintf(w, "class %sBaseClass;\n", cppClassPrefix)
	fmt.Fprintf(w, "class %sWrapper;\n", cppClassPrefix)
	for i := 0; i < len(component.Classes); i++ {
		class := component.Classes[i]
		fmt.Fprintf(w, "class %s%s;\n", cppClassPrefix, class.ClassName)
	}

	fmt.Fprintf(w, "\n")

	fmt.Fprintf(w, "/*************************************************************************************************************************\n")
	fmt.Fprintf(w, " Declaration of shared pointer types \n")
	fmt.Fprintf(w, "**************************************************************************************************************************/\n")

	fmt.Fprintf(w, "\n")

	fmt.Fprintf(w, "typedef std::shared_ptr<%sBaseClass> P%sBaseClass;\n", cppClassPrefix, NameSpace)
	fmt.Fprintf(w, "typedef std::shared_ptr<%sWrapper> P%sWrapper;\n", cppClassPrefix, NameSpace)
	for i := 0; i < len(component.Classes); i++ {
		class := component.Classes[i]
		fmt.Fprintf(w, "typedef std::shared_ptr<%s%s> P%s%s;\n", cppClassPrefix, class.ClassName, NameSpace, class.ClassName)
	}

	fmt.Fprintf(w, "     \n")
	fmt.Fprintf(w, "/*************************************************************************************************************************\n")
	fmt.Fprintf(w, " Class E%sException \n", NameSpace)
	fmt.Fprintf(w, "**************************************************************************************************************************/\n")
	fmt.Fprintf(w, "class E%sException : public std::runtime_error {\n", NameSpace)
	fmt.Fprintf(w, "  protected:\n")
	fmt.Fprintf(w, "    /**\n")
	fmt.Fprintf(w, "    * Error code for the Exception.\n")
	fmt.Fprintf(w, "    */\n")
	fmt.Fprintf(w, "    %sResult m_errorcode;\n", NameSpace)
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "  public:\n")
	fmt.Fprintf(w, "    /**\n")
	fmt.Fprintf(w, "    * Exception Constructor.\n")
	fmt.Fprintf(w, "    */\n")
	fmt.Fprintf(w, "    E%sException (%sResult errorcode)\n", NameSpace, NameSpace)
	fmt.Fprintf(w, "        : std::runtime_error (\"%s Error \" + std::to_string (errorcode))\n", NameSpace)
	fmt.Fprintf(w, "    {\n")
	fmt.Fprintf(w, "        m_errorcode = errorcode;\n")
	fmt.Fprintf(w, "    }\n")
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "    /**\n")
	fmt.Fprintf(w, "    * Returns error code\n")
	fmt.Fprintf(w, "    */\n")
	fmt.Fprintf(w, "    %sResult getErrorCode ()\n", NameSpace)
	fmt.Fprintf(w, "    {\n")
	fmt.Fprintf(w, "        return m_errorcode;\n")
	fmt.Fprintf(w, "    }\n")
	fmt.Fprintf(w, "\n")

	fmt.Fprintf(w, "};\n")

	fmt.Fprintf(w, "     \n")
	
	
	fmt.Fprintf(w, "/*************************************************************************************************************************\n")
	fmt.Fprintf(w, " Class %sWrapper \n", cppClassPrefix)
	fmt.Fprintf(w, "**************************************************************************************************************************/\n")

	fmt.Fprintf(w, "class %sWrapper {\n", cppClassPrefix)
	fmt.Fprintf(w, "public:\n")
	
	fmt.Fprintf(w, "    \n")
	fmt.Fprintf(w, "    %sWrapper (const std::string &sFileName)\n", cppClassPrefix)
	fmt.Fprintf(w, "    {\n")
	fmt.Fprintf(w, "        CheckError (nullptr, initWrapperTable (&m_WrapperTable));\n")
	fmt.Fprintf(w, "        CheckError (nullptr, loadWrapperTable (&m_WrapperTable, sFileName.c_str ()));\n")
	fmt.Fprintf(w, "    }\n")
	fmt.Fprintf(w, "    \n")
	
	fmt.Fprintf(w, "    static P%sWrapper loadLibrary (const std::string &sFileName)\n", NameSpace)	
	fmt.Fprintf(w, "    {\n");
	fmt.Fprintf(w, "        return std::make_shared<%sWrapper> (sFileName);\n", cppClassPrefix);
	fmt.Fprintf(w, "    }\n")
	fmt.Fprintf(w, "    \n")
	

	fmt.Fprintf(w, "    ~%sWrapper ()\n", cppClassPrefix)
	fmt.Fprintf(w, "    {\n")
	fmt.Fprintf(w, "        releaseWrapperTable (&m_WrapperTable);\n")
	fmt.Fprintf(w, "    }\n")
	fmt.Fprintf(w, "    \n")
	
	fmt.Fprintf(w, "    void CheckError(%sHandle handle, %sResult nResult)\n", NameSpace, NameSpace)
	fmt.Fprintf(w, "    {\n")
	fmt.Fprintf(w, "        if (nResult != 0) \n")
	fmt.Fprintf(w, "            throw E%sException (nResult);\n", NameSpace)
	fmt.Fprintf(w, "    }\n")
	fmt.Fprintf(w, "    \n")
	
	fmt.Fprintf(w, "\n")

	for j := 0; j < len(global.Methods); j++ {
		method := global.Methods[j]

		err := writeDynamicCPPMethodDefinition(method, w, NameSpace, "Wrapper", true)
		if err != nil {
			return err
		}
	}
		
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "private:\n")
	fmt.Fprintf(w, "    s%sDynamicWrapperTable m_WrapperTable;\n", NameSpace)
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "    %sResult initWrapperTable (s%sDynamicWrapperTable * pWrapperTable);\n", NameSpace, NameSpace)
	fmt.Fprintf(w, "    %sResult releaseWrapperTable (s%sDynamicWrapperTable * pWrapperTable);\n", NameSpace, NameSpace)
	fmt.Fprintf(w, "    %sResult loadWrapperTable (s%sDynamicWrapperTable * pWrapperTable, const char * pLibraryFileName);\n", NameSpace, NameSpace)
	fmt.Fprintf(w, "\n")
	for i := 0; i < len(component.Classes); i++ {

		class := component.Classes[i]
		cppClassName := cppClassPrefix + class.ClassName
		fmt.Fprintf(w, "    friend class %s;\n", cppClassName)
		
	}
	fmt.Fprintf(w, "\n")	
	fmt.Fprintf(w, "};\n")
	fmt.Fprintf(w, "\n")
	
	
	fmt.Fprintf(w, "/*************************************************************************************************************************\n")
	fmt.Fprintf(w, " Class %sBaseClass \n", cppClassPrefix)
	fmt.Fprintf(w, "**************************************************************************************************************************/\n")

	fmt.Fprintf(w, "class %sBaseClass {\n", cppClassPrefix)
	fmt.Fprintf(w, "  protected:\n")
	fmt.Fprintf(w, "    /* Wrapper Object that created the class..*/\n")
	fmt.Fprintf(w, "    %sWrapper * m_pWrapper;\n", cppClassPrefix)
	fmt.Fprintf(w, "    /* Handle to Instance in library*/\n")
	fmt.Fprintf(w, "    %sHandle m_pHandle;\n", NameSpace)
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "    /* Checks for an Error code and raises Exceptions */\n")
	fmt.Fprintf(w, "    void CheckError(%sResult nResult)\n", NameSpace)
	fmt.Fprintf(w, "    {\n")
	fmt.Fprintf(w, "        if (m_pWrapper != nullptr)\n")
	fmt.Fprintf(w, "            m_pWrapper->CheckError (m_pHandle, nResult);\n")
	fmt.Fprintf(w, "    }\n")
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "  public:\n")
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "    /**\n")
	fmt.Fprintf(w, "    * %sBaseClass::%sBaseClass - Constructor for Base class.\n", cppClassPrefix, cppClassPrefix)
	fmt.Fprintf(w, "    */\n")
	fmt.Fprintf(w, "    %sBaseClass(%sWrapper * pWrapper, %sHandle pHandle)\n", cppClassPrefix, cppClassPrefix, NameSpace)
	fmt.Fprintf(w, "        : m_pWrapper (pWrapper), m_pHandle (pHandle)\n")
	fmt.Fprintf(w, "    {\n")
	fmt.Fprintf(w, "    }\n")
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "    /**\n")
	fmt.Fprintf(w, "    * %sBaseClass::~%sBaseClass - Destructor for Base class.\n", cppClassPrefix, cppClassPrefix)
	fmt.Fprintf(w, "    */\n")

	fmt.Fprintf(w, "    virtual ~%sBaseClass()\n", cppClassPrefix)
	fmt.Fprintf(w, "    {\n")
	fmt.Fprintf(w, "        if (m_pWrapper != nullptr)\n")
	fmt.Fprintf(w, "            m_pWrapper->%s (this);\n", component.Global.ReleaseMethod)
	fmt.Fprintf(w, "        m_pWrapper = nullptr;\n")
	fmt.Fprintf(w, "    }\n")
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "    /**\n")
	fmt.Fprintf(w, "    * %sBaseClass::GetHandle - Returns handle to instance.\n", cppClassPrefix)
	fmt.Fprintf(w, "    */\n")
	fmt.Fprintf(w, "    %sHandle GetHandle()\n", NameSpace)
	fmt.Fprintf(w, "    {\n")
	fmt.Fprintf(w, "        return m_pHandle;\n")
	fmt.Fprintf(w, "    }\n")
	fmt.Fprintf(w, "};\n")
	
	fmt.Fprintf(w, "     \n")
	
	for i := 0; i < len(component.Classes); i++ {

		class := component.Classes[i]
		cppClassName := cppClassPrefix + class.ClassName

		parentClassName := class.ParentClass
		if parentClassName == "" {
			parentClassName = "BaseClass"
		}
		cppParentClassName := cppClassPrefix + parentClassName

		fmt.Fprintf(w, "     \n")
		fmt.Fprintf(w, "/*************************************************************************************************************************\n")
		fmt.Fprintf(w, " Class %s \n", cppClassName)
		fmt.Fprintf(w, "**************************************************************************************************************************/\n")
		fmt.Fprintf(w, "class %s : public %s {\n", cppClassName, cppParentClassName)
		fmt.Fprintf(w, "  public:\n")
		fmt.Fprintf(w, "    \n")
		fmt.Fprintf(w, "    /**\n")
		fmt.Fprintf(w, "    * %s::%s - Constructor for %s class.\n", cppClassName, cppClassName, class.ClassName)
		fmt.Fprintf(w, "    */\n")
		fmt.Fprintf(w, "    %s (%sWrapper * pWrapper, %sHandle pHandle)\n", cppClassName, cppClassPrefix, NameSpace)
		fmt.Fprintf(w, "        : %s (pWrapper, pHandle)\n", cppParentClassName);
		fmt.Fprintf(w, "    {\n")
		fmt.Fprintf(w, "    }\n")
		fmt.Fprintf(w, "    \n")

		for j := 0; j < len(class.Methods); j++ {
			method := class.Methods[j]

			err := writeDynamicCPPMethod(method, w, NameSpace, class.ClassName, false)
			if err != nil {
				return err
			}

		}

		fmt.Fprintf(w, "};\n\n")

	}
	

	for j := 0; j < len(global.Methods); j++ {
		method := global.Methods[j]

		err := writeDynamicCPPMethod(method, w, NameSpace, "Wrapper", true)
		if err != nil {
			return err
		}
	}

	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "    inline %sResult %sWrapper::initWrapperTable (s%sDynamicWrapperTable * pWrapperTable)\n", NameSpace, cppClassPrefix, NameSpace)
	fmt.Fprintf(w, "    {\n")
	
	buildDynamicCInitTableCode (component, w, NameSpace, BaseName, "        ");
	
	fmt.Fprintf(w, "    }\n")
	fmt.Fprintf(w, "\n")

	fmt.Fprintf(w, "    inline %sResult %sWrapper::releaseWrapperTable (s%sDynamicWrapperTable * pWrapperTable)\n", NameSpace, cppClassPrefix, NameSpace)
	fmt.Fprintf(w, "    {\n")

	buildDynamicCReleaseTableCode (component, w, NameSpace, BaseName, "        ", "initWrapperTable");

	fmt.Fprintf(w, "        \n")
	fmt.Fprintf(w, "    }\n")
	fmt.Fprintf(w, "\n")
	
	fmt.Fprintf(w, "    inline %sResult %sWrapper::loadWrapperTable (s%sDynamicWrapperTable * pWrapperTable, const char * pLibraryFileName)\n", NameSpace, cppClassPrefix, NameSpace)
	fmt.Fprintf(w, "    {\n")
	
	buildDynamicCLoadTableCode (component, w, NameSpace, BaseName, "        ");
	
	fmt.Fprintf(w, "    }\n")
	fmt.Fprintf(w, "\n")
	
	fmt.Fprintf(w, "} // namespace %s\n", NameSpace)
	fmt.Fprintf(w, "\n")
	
	fmt.Fprintf(w, "#endif // __%s_DYNAMICCPPHEADER\n", strings.ToUpper(NameSpace))
	fmt.Fprintf(w, "\n")

	return nil
}



// BuildBindingCppDynamic builds dynamic headeronly C++-bindings of a library's API in form of dynamically loaded functions
// handles.
func BuildBindingCppDynamic(component ComponentDefinition, outputFolder string) error {

	namespace := component.NameSpace;
	libraryname := component.LibraryName;
	baseName := component.BaseName;

	DynamicCHeader := path.Join(outputFolder, baseName+"_dynamic.h");
	log.Printf("Creating \"%s\"", DynamicCHeader)
	dynhfile, err := os.Create(DynamicCHeader)
	if err != nil {
		return err;
	}
	WriteLicenseHeader(dynhfile, component,
		fmt.Sprintf("This is an autogenerated plain C Header file in order to allow an easy\n use of %s", libraryname),
		true)
	err = buildDynamicCHeader(component, dynhfile, namespace, baseName, true)
	if err != nil {
		return err;
	}
	
	DynamicCppHeader := path.Join(outputFolder, baseName+"_dynamic.hpp");
	log.Printf("Creating \"%s\"", DynamicCppHeader)
	dynhppfile, err := os.Create(DynamicCppHeader)
	if err != nil {
		return err;
	}
	WriteLicenseHeader(dynhppfile, component,
		fmt.Sprintf("This is an autogenerated C++ Header file in order to allow an easy\n use of %s", libraryname),
		true)
	err = buildDynamicCppHeader(component, dynhppfile, namespace, baseName)
	if err != nil {
		return err;
	}
	
	return nil;
	
	
}