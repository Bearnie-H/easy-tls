#!/usr/bin/env bash

#   Author: Joseph Sadden
#   Date:   24th June, 2020

#   Purpose:
#
#       This script will copy over one of the plugin templates to a new project folder.

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

HelpFlag=

EasyTLSRepository=

#   Command-line / Global variables
if [ -z "$GOPATH" ]; then
    log $LOG_ERROR "\$GOPATH not set, defaulting to \$HOME/go"
    if [ -z "$HOME" ]; then
        log $LOG_FATAL "\$HOME not set!"
    else
        EasyTLSRepository=="$HOME/go"
    fi
else
    EasyTLSRepository="$GOPATH/"
fi
EasyTLSRepository="$EasyTLSRepository/src/github.com/Bearnie-H/easy-tls"

ServerType="server"
ClientType="client"
TemplateFolder=

PluginType=
PluginName=
DestinationFolder=

#   Function to display a help/usage menu to the user in a standardized format.
function helpMenu() {

    #   Get the name of the executable
    local scriptName=$(basename "$0")

    #   Print the current help menu/usage information to the user
    echo -e "
    $BLUE$scriptName   -   A bash tool to create new EasyTLS-Compliant plugins from the templates of the EasyTLS repository.$WHITE

    $GREEN$scriptName$YELLOW { -c | -s } -d Save-Folder -n Plugin-Name [-h] [-o Output-File] [-q] [-z]$WHITE

    "$YELLOW"Save Options:$WHITE
        $BLUE-d$WHITE  -    Save Folder. The full path to the folder to save the module at.
        $BLUE-n$WHITE  -    Save Name. The name to assign to the plugin.

    "$YELLOW"Module Types:$WHITE
        $BLUE-c$WHITE  -    Client Flag. Create a Client plugin from the template.
        $BLUE-s$WHITE  -    Server Flag. Create a Server plugin from the template.

    "$YELLOW"Output Options:$WHITE
        $BLUE-o$WHITE  -    Log File. Redirect STDOUT to the given file, creating it if it doesn't exist yet.
        $BLUE-q$WHITE  -    Quiet mode. Only print out fatal errors and suppress all other output.
        $BLUE-z$WHITE  -    Raw Mode. Disables colouring, useful when the ANSI escape codes would be problematic.

    "$YELLOW"Miscellaneous Options:$WHITE
        $BLUE-h$WHITE  -    Help Menu. Display this help menu and exit.

    "$GREEN"Note:$WHITE
        <...>
    "$CLEAR
}

function cleanup() {

    #   Implement whatever cleanup logic is needed for the specific script, followed by resetting the terminal and exiting.

    # if [ $1 -eq 0 ]; then
    #     log $LOG_INFO "Successfully executed and beginning cleanup..."
    # else
    #     log $LOG_ATTENTION "Unsuccessfully executed and beginning cleanup..."
    # fi

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

    assertDirectoryExists "$DestinationFolder"
    if ! directoryExists "$TemplateFolder"; then
        log $LOG_FATAL "Failed to find EasyTLS [ $PluginType ] module template folder [ $TemplateFolder ]."
        return
    fi

    if ! cp -rp "$TemplateFolder" "$DestinationFolder/$PluginName"; then
        log $LOG_FATAL "Failed to create [ $PluginType ] module from template!"
        return
    else
        log $LOG_INFO "Successfully created new [ $PluginType ] module at [ $DestinationFolder/$PluginName ]."
    fi

    log $LOG_INFO "Setting \"PluginName\" field of [ $PluginType ] module..."
    if ! sed -i "s/^\(.*\)DEFINE_ME\(.*\)$/\1$PluginName\2/" $DestinationFolder/$PluginName/module-definitions.go; then
        log $LOG_ERROR "Failed to set \"PluginName\" field of [ $PluginType ] module!"
    else
        log $LOG_NOTICE "Successfully to set \"PluginName\" field of [ $PluginType ] module to [ $PluginName ]."
    fi

    return
}


#   Parse the command line arguments.  Add the flag name to the list (in alphabetical order), and add a ":" after if it requires an argument present.
#   The value of the argument will be located in the "$OPTARG" variable
while getopts "cd:hn:o:qsz" opt; do
    case "$opt" in
    c)  PluginType="$ClientType"
        ;;
    d)  DestinationFolder="$OPTARG"
        ;;
    h)  HelpFlag=1
        ;;
    n)  PluginName="$OPTARG"
        ;;
    o)  OutputFile="$OPTARG"
        ;;
    q)  Quiet="-q"
        ;;
    s)  PluginType="$ServerType"
        ;;
    z)  ColourLog=
        ColourFlag="-z"
        ;;
    \?) HelpFlag=2
        ;;
    esac
done

case $HelpFlag in
    1)  helpMenu
        cleanup 0
        ;;
    2)  helpMenu
        cleanup 1
        ;;
esac

argSet "$OutputFile" "-o"

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

#   Assert all of the required arguments are set here

argSet "$PluginType" "{ -c | -s }"
argSet "$DestinationFolder" "-d"
argSet "$PluginName" "-n"

TemplateFolder="$EasyTLSRepository/examples/example-$PluginType-plugin"

#   Call main, running the full logic of the script.
main

cleanup 0
