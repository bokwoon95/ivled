#!/bin/bash

# lapi core API
lapi_modules_taken() {
  curl "https://ivle.nus.edu.sg/api/Lapi.svc/Modules_Taken?APIKey=$LAPIkey&AuthToken=$AuthToken&StudentID=$StudentID" | jq .
}

lapi_modules_search() {
  # $1 - ModuleCode e.g. MA1513
  curl "https://ivle.nus.edu.sg/api/Lapi.svc/Modules_Search?APIKey=$LAPIkey&AuthToken=$AuthToken&IncludeAllInfo=false&ModuleCode=$1" | jq .
}

lapi_workbins() {
  # $1 - CourseID e.g. 08794bbd-b65a-4389-ad06-078b09fc729e
  curl "https://ivle.nus.edu.sg/api/Lapi.svc/Workbins?APIKey=$LAPIkey&AuthToken=$AuthToken&CourseID=$1" | jq .
}

lapi_downloadfile() {
  # $1 - CourseID e.g. 5444db22-b035-406a-9c46-2cdac6e30bd3
  curl "https://ivle.nus.edu.sg/api/downloadfile.ashx?APIKey=$LAPIkey&AuthToken=$AuthToken&ID=$1&target=workbin" | jq .
}

modules_search() {
  regex="($1/)|(/$1)|(^$1\$)"
  curl "https://ivle.nus.edu.sg/api/Lapi.svc/Modules_Search?APIKey=$LAPIkey&AuthToken=$AuthToken&IncludeAllInfo=false&ModuleCode=$1"\
    | jq --arg regex "$regex" '.Results[] | select(.CourseAcadYear == "2018/2019" and .CourseSemester == "Semester 2" and (.CourseCode | test($regex; "ix")))'\
    | jq '{
  ID: .ID,
  CourseCode: .CourseCode,
  CourseName: .CourseName,
  CourseAcadYear: .CourseAcadYear,
  CourseSemester: .CourseSemester}'
}

modules_search_fuzzy() {
  curl "https://ivle.nus.edu.sg/api/Lapi.svc/Modules_Search?APIKey=$LAPIkey&AuthToken=$AuthToken&IncludeAllInfo=false&ModuleCode=$1"\
    | jq '.Results[] | select(.CourseAcadYear == "2018/2019" and .CourseSemester == "Semester 2")'\
    | jq '{
  ID: .ID,
  CourseCode: .CourseCode,
  CourseName: .CourseName,
  CourseAcadYear: .CourseAcadYear,
  CourseSemester: .CourseSemester}'\
    | jq -s .
}

old_modules_taken() {
  if [[ -f modules_taken.json && "$1" == "" ]]; then
    < modules_taken.json jq .
  else
    echo "======================================"
    echo "cURL-ing your modules this semester..."
    echo "======================================"
    curl "https://ivle.nus.edu.sg/api/Lapi.svc/Modules_Taken?APIKey=$LAPIkey&AuthToken=$AuthToken&StudentID=$StudentID"\
      | jq '.Results[] | select(.AcadYear == "2018/2019" and .Semester == "2")'\
      | jq '{
    ModuleCode: .ModuleCode,
    ModuleTitle: .ModuleTitle}'\
      | jq -s .\
      | tee modules_taken.json
  fi
}

modules_taken() {
  if [[ -f modules_taken.json ]]; then
    echo "======================================"
    echo "cURL-ing your modules this semester..."
    echo "======================================"
    json_input=$(curl "https://ivle.nus.edu.sg/api/Lapi.svc/Modules_Taken?APIKey=$LAPIkey&AuthToken=$AuthToken&StudentID=$StudentID"\
      | jq '.Results[] | select(.AcadYear == "2018/2019" and .Semester == "2")'\
      | jq '{
    ModuleCode: .ModuleCode,
    ModuleTitle: .ModuleTitle}'\
      | jq -s)
    echo "======================"
    echo "Initial cURL completed"
    echo "======================"

    i=0
    for item in $(echo $json_input | jq -r '.[] | @base64'); do
      ModuleCode=$(echo $item | base64 --decode | jq -r .ModuleCode)
      regex="($ModuleCode/)|(/$ModuleCode)|(^$ModuleCode\$)"
      ID=$(curl "https://ivle.nus.edu.sg/api/Lapi.svc/Modules_Search?APIKey=$LAPIkey&AuthToken=$AuthToken&IncludeAllInfo=false&ModuleCode=$ModuleCode"\
        | jq --arg regex "$regex" '.Results[] | select(.CourseAcadYear == "2018/2019" and .CourseSemester == "Semester 2" and (.CourseCode | test($regex; "ix")))'\
        | jq -r '.ID')
      echo $ID
      echo $item | base64 --decode | jq --arg id $ID '.+{ID: $id}' | tee temp$i.json
      i=$((i+1))
    done
    jq -s . temp*.json | tee modules_taken.json
    rm temp*.json
  else
  fi
}

iniread() {
  inifile=$1
  section=$2
  key=$3
  sed -n "/\[$section\]/,/\[/{/^(\W|)$key=/p}" $inifile
}

iniwrite() {
  inifile=$1
  section=$2
  key=$3
  value=$4
  perl -p00 -e "/$section/ && s/($key.*=)\s*\d/$1 $value/" $inifile
}
