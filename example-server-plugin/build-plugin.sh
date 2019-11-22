#!/usr/bin/env bash

#   Terminal Colour Codes
RED="\e[91m"
GREEN="\e[92m"
YELLOW="\e[93m"
BLUE="\e[96m"
CLEAR="\e[0m"

artefactDirectory="$GOPATH/src/easytls-compiled-plugins"
Force=
ModuleType=

function helpMenu() {
    scriptName=$(basename "$0")
    echo -e "
    $GREEN$scriptName - The standard tool for building a Golang plugin for the EasyTLS framework.$CLEAR
    
    $scriptName  $YELLOW[-fh]$CLEAR
    
        $BLUE-f$CLEAR  Force Mode  -   This will force building the plugin,
                            even if the module is not marked for
                            release with the presence of a \"RELEASE\"
                            file within the module folder
        $BLUE-h$CLEAR  Help Menu   -   Print this help menu and exit
        "
    exit 0
}

function incrementBuildCount() {
    if [ ! -f version.go ]; then
        echo -e $RED"ERROR! Failed to find required \"version.go\" file in $(pwd)!"$CLEAR
        exit 1
    fi
    local lastBuildCount=$(cat version.go | grep 'Build' | sed 's/[ \t,]*//g' | awk -F: '{print $2}');
    lastBuildCount=$(( $lastBuildCount + 1 ))
    sed -i "s/Build\:[ \t]*[0-9]*/Build:$lastBuildCount/g" version.go 
    gofmt -s -w version.go
}

function ShouldRun() {

    read -p "Build $ModuleType Module \"$(basename $(pwd)\")? [y/N]: " response

    if [[ "$response" =~ [yY] ]]; then
        echo "Yes"
    fi
}

function main() {

    echo -e $BLUE"\nExecuting $0"$CLEAR

    buildDir=$(cd "$(dirname "$0")"; pwd)

    cd "$buildDir"

    if [ ! -z $(echo "$buildDir" | grep "client") ]; then
        ModuleType="Client"
    elif [ ! -z $(echo "$buildDir" | grep "server") ]; then
        ModuleType="Server"
    else
        ModuleType="Generic"
    fi

    if [ -z "$Force" ]; then
        if [ ! -e "RELEASE" ]; then
            echo -e $YELLOW"WARNING: Module $(basename $buildDir) not marked for release. Skipping..."$CLEAR
            exit 0
        else
            shouldRun=$(ShouldRun)
            if [ -z "$shouldRun" ]; then
                echo -e $BLUE"Skipping build..."$CLEAR
                exit 0
            fi
            echo -e $GREEN"Building Module $(basename $buildDir)..."$CLEAR
        fi
    fi

    mkdir -p "$artefactDirectory"
    echo -e $BLUE"Updating Module Build Number..."$CLEAR
    incrementBuildCount
    echo -e $BLUE"Finished Updating Module Build Number."$CLEAR

    pluginName=$(basename "$buildDir")
    pluginName=$(echo "$ModuleType" | tr '[:upper:]' '[:lower:]')"$pluginName"

    go build -buildmode=plugin -o "$artefactDirectory/$pluginName.so" -gcflags="all=-N -l"
    if [ $? -eq 0 ]; then
        echo -e $GREEN"Built Module $(basename $buildDir). Artefact located at \"$artefactDirectory/$pluginName.so\" "$CLEAR
        cd - > /dev/null
    else
        echo -e $RED"ERROR! Failed to Build Module $(basename $buildDir)."$CLEAR
        cd - > /dev/null
        exit 1
    fi

    echo -e $BLUE"Finished Executing $0\n"$CLEAR
}

#   Parse the command-line arguments
while getopts "fh" opt; do
    case "$opt" in
    f)  Force="1"
        ;;
    h)  helpMenu
        ;;
    \?) helpMenu
        ;;
    esac
done

main
