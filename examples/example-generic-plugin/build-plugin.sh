#!/usr/bin/env bash

#   Author: 
#   Date:   

#   Purpose:
#
#       This is the standard Build Script for an EasyTLS-compliant Golang module.

#   Terminal Colour Codes
FATAL="\e[7m\e[31m"
RED="\e[91m"
GREEN="\e[92m"
YELLOW="\e[93m"
AMBER="\e[33m"
BLUE="\e[96m"
WHITE="\e[97m"
CLEAR="\e[0m"

# Global Logging variables
Quiet=
OutputFile="-"  #    "-" indicates STDOUT
ColourLog=1  #  Flag for adding colours to the log, empty means no colouring
ColourFlag=
LogPrefix=
LOG_INFO=0
LOG_NOTICE=1
LOG_ATTENTION=2
LOG_WARNING=3
LOG_ERROR=4
LOG_FATAL=5
LogLevels=( "INFO:  " "NOTE:  " "ATTN:  " "WARN:  " "ERROR: " "FATAL: ")
LogColours=("$WHITE"  "$BLUE"   "$YELLOW" "$AMBER"  "$RED"    "$FATAL")
TimeFormat="+%F %H:%M:%S"

#   Catch errors, interrupts, and more to allow a safe shutdown
trap 'cleanup 1' 1 2 3 6 8 9 15

#   Command-line / Global variables
HelpFlag=
BuildHash=
RaceDetector=""

artefactDirectory="$GOPATH/src/easytls-compiled-plugins"
ForceMode=

#   Function to display a help/usage menu to the user in a standardized format.
function helpMenu() {

    #   Get the name of the executable
    local scriptName=$(basename "$0")

    #   Print the current help menu/usage information to the user
    echo -e "
    $BLUE$scriptName   -   A bash tool to build an EasyTLS-Compliant Golang module, either server or client.$WHITE

    $GREEN$scriptName$YELLOW -x Hash [-f] [-h] [-o Output-File] [-q] [-r] [-z] $WHITE

    "$YELLOW"Build Options:$WHITE
        $BLUE-f$WHITE  -    Force Mode. Build the module, even without the presence of a RELEASE file.
        $BLUE-r$WHITE  -    Race Detection. Include the Golang Race Detector in the binary.
                This comes at a slight memory and performance cost, but can be valuable in testing.

    "$YELLOW"Output Options:$WHITE
        $BLUE-o$WHITE  -    Log File. Redirect STDOUT to the given file, creating it if it doesn't exist yet.
        $BLUE-q$WHITE  -    Quiet mode. Only print out fatal errors and suppress all other output.
        $BLUE-z$WHITE  -    Raw Mode. Disables colouring, useful when the ANSI escape codes would be problematic.

    "$YELLOW"Miscellaneous Options:$WHITE
        $BLUE-h$WHITE  -    Help Menu. Display this help menu and exit.
        $BLUE-x$WHITE  -    Build Hash. Used to allow concurrent builds, and to prevent clobbering of prior builds.

    "$GREEN"Note:$WHITE
        The flags passed to the \"go build...\" command are very important, as these must be present here
        and they must be present in the Plugin build scripts to allow the Plugins to be successfully
        loaded into the framework during execution.  Furthermore, it is best practice to compile the
        framework application AND the desired plugins in one pass, as the ABI of the compiled Shared Object
        Libraries may vary if built with different versions of the imported modules.
    "$CLEAR
}

function cleanup() {

    #   Implement whatever cleanup logic is needed for the specific script, followed by resetting the terminal and exiting.
    #   ...

    if [ $1 -eq 0 ]; then
        log $LOG_INFO "Successfully executed and beginning cleanup..."
    else
        log $LOG_ATTENTION "Unsuccessfully executed and beginning cleanup..."
    fi

    stop $1
}

function stop() {
    exit $1
}

function SetLogPrefix() {
    LogPrefix="$1"
}

#   $1 -> Log Level
#   $2 -> Log Message
function log() {

    local Level=$1

    #   Only log if not in quiet mode, or it's a fatal error
    if [[ -z "$Quiet" ]] || [[ $Level -eq $LOG_FATAL ]]; then

        local Message="$2"
        local Timestamp="[$(date "$TimeFormat")]"

        local ToWrite=

        if [ -z "$LogPrefix" ]; then
            ToWrite="$Timestamp ${LogLevels[$Level]} $Message"
        else
            ToWrite="$Timestamp [ $LogPrefix ] ${LogLevels[$Level]}: $Message"
        fi

        #   If log colouring is on, check if it's writing to an output file
        if [ ! -z "$ColourLog" ] && [[ "$OutputFile" == "-" ]]; then
            ToWrite="${LogColours[$Level]}""$ToWrite""$CLEAR"
        fi

        #   Attention and higher should be logged to STDERR, Info and Notice to STDOUT
        if [ $Level -ge $LOG_ATTENTION ]; then
            echo -e "$ToWrite" >&2
        else
            if [[ "$OutputFile" == "-" ]]; then
                echo -e "$ToWrite" >&1
            else
                echo -e "$ToWrite" >> "$OutputFile"
            fi
        fi

        #   If it's a fatal error, full exit
        if [ $Level -eq $LOG_FATAL ]; then
            cleanup 1
        fi
    fi
}

#   Helper function to allow asserting that required arguments are set.
function argSet() {
    local argToCheck="$1"
    local argName="$2"

    if [ -z "$argToCheck" ]; then
        log $LOG_FATAL "Required argument [ $argName ] not set!"
    fi
}

#   Helper function to allow checking for the existence of files on disk.
function fileExists() {
    local FilenameToCheck="$1"

    if [ ! -f "$FilenameToCheck" ]; then
        log $LOG_ATTENTION "File [ $FilenameToCheck ] does not exist."
        return 1
    fi

    return 0
}

#   Helper function to allow checking for the existence of directories on disk.
function directoryExists() {
    local DirectoryToCheck="$1"

    if [ ! -d "$DirectoryToCheck" ]; then
        log $LOG_ATTENTION "Directory [ $DirectoryToCheck ] does not exist."
        return 1
    fi

    return 0
}

#   Helper function to either assert that a given directory does exist (creating it if necessary) or exiting if it cannot.
function assertDirectoryExists() {

    local DirectoryToCheck="$1"

    if ! directoryExists "$DirectoryToCheck"; then
        if ! mkdir -p "$DirectoryToCheck"; then
            log $LOG_FATAL "Failed to create directory [ $DirectoryToCheck ]!"
        fi

        log $LOG_NOTICE "Successfully created directory [ $DirectoryToCheck ]."
    fi
}

#   Main function, this is the entry point of the actual logic of the script, AFTER all of the input validation and top-level script pre-script set up has been completed.
function main() {

    artefactDirectory="$artefactDirectory-$BuildHash"

    buildDir=$(cd "$(dirname "$0")"; pwd)
    cd "$buildDir"

    moduleName=$(basename "$buildDir")
    if [ ! -z $(echo "$buildDir" | grep 'server') ]; then
        moduleName="server-$moduleName"
    else
        moduleName="client-$moduleName"
    fi

    if [ -z "$ForceMode" ]; then
        if [ ! -e "RELEASE" ]; then
            log $LOG_ATTENTION "Module $(basename $buildDir) not marked for release. Skipping..."
            exit 0
        else
            log $LOG_NOTICE "Building Module $(basename $buildDir)..."
        fi
    fi

    checkGitBranch "$moduleName"

    if [ ! -z "$RaceDetector" ]; then
        log $LOG_NOTICE "Building [ with ] Race Detection."
    else
        log $LOG_NOTICE "Building [ without ] Race Detection."
    fi

    log $LOG_INFO "Updating Module Build Number..."
    incrementBuildCount

    log $LOG_INFO "Compiling Golang Module [ $moduleName ]..."
    go build -buildmode=plugin -o "$artefactDirectory/$moduleName.so" -gcflags="all=-N -l" "$RaceDetector"
    if [ $? -eq 0 ]; then
        log $LOG_NOTICE "Successfully built module [ $moduleName ]."
        log $LOG_NOTICE "Build artefact located at [ $artefactDirectory/$moduleName.so ]"
        cd - 2>&1 > /dev/null
    else
        log $LOG_FATAL "Failed to build module [ $moduleName ]"
        cd - 2>&1 > /dev/null
    fi

    return
}

function incrementBuildCount() {

    local VersionFile="version.go"

    if ! fileExists "$VersionFile"; then
        log $LOG_FATAL "Cannot find required file [ $VersionFile ]!"
    fi

    local lastBuildCount=$(cat "$VersionFile" | grep 'Build' | sed 's/[ \t,]*//g' | awk -F: '{print $2}');
    lastBuildCount=$(( $lastBuildCount + 1 ))
    sed -i "s/Build\:[ \t]*[0-9]*/Build:$lastBuildCount/g" "$VersionFile"
    gofmt -s -w "$VersionFile"
}

function checkGitBranch() {

    #   We want to explicitly let the user know which branch they are on when building the application.
    local CurrentBranch=$(git branch | grep '\*' | sed 's/^\* \(.*\)$/\1/g')

    local moduleName="$1"
    
    log $LOG_INFO "Building Golang module [ $moduleName ] from branch [ $CurrentBranch ]."
}

#   Parse the command line arguments.  Add the flag name to the list (in alphabetical order), and add a ":" after if it requires an argument present.
#   The value of the argument will be located in the "$OPTARG" variable
while getopts "fho:qrx:z" opt; do
    case "$opt" in
    f)  ForceMode="-f"
        ;;
    h)  HelpFlag=1
        ;;
    o)  OutputFile="$OPTARG"
        ;;
    q)  Quiet="-q"
        ;;
    r)  RaceDetector="-race"
        ;;
    x)  BuildHash="$OPTARG"
        ;;
    z)  ColourLog=
        ColourFlag="-z"
        ;;
    \?) HelpFlag=2
        ;;
    esac
done

case HelpFlag in
    1)  helpMenu
        cleanup 0
        ;;
    2)  helpMenu
        cleanup 1
        ;;
esac

SetLogPrefix "$(basename "$(dirname "$(readlink -e "$0")")")"

if [[ ! "$OutputFile" == "-" ]]; then

    #   Only assert this here, in case multiple -o arguments are given.
    #   Only create the file of the final argument.
    assertDirectoryExists "$(dirname "$OutputFile")"

    if ! fileExists "$OutputFile"; then
        #   Create the empty file.
        >"$OutputFile"
    fi

    #   Resolve the output file to an absolute path
    OutputFile=$(readlink -e "$OutputFile")
fi

argSet "$BuildHash" "-x"

#   Call main, running the full logic of the script.
main

cleanup 0
