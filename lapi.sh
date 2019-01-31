#!/bin/bash

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
  curl "https://ivle.nus.edu.sg/api/Lapi.svc/Modules_Search?APIKey=$LAPIkey&AuthToken=$AuthToken&IncludeAllInfo=false&ModuleCode=$1"\
    | jq '.Results[] |
    {
      ID: .ID,
      CourseCode: .CourseCode,
      CourseName: .CourseName,
      CourseAcadYear: .CourseAcadYear,
      CourseSemester: .CourseSemester
    }'
}

modules_taken() {
  if [[ -f modules_taken.json ]]; then
    < modules_taken.json jq .
  else
    echo "=============================================="
    echo "modules_taken.json does not exist, cURL-ing..."
    echo "=============================================="
    curl "https://ivle.nus.edu.sg/api/Lapi.svc/Modules_Taken?APIKey=$LAPIkey&AuthToken=$AuthToken&StudentID=$StudentID"\
      | jq '.Results[] | select(.AcadYear == "2018/2019" and .Semester == "2")'\
      | jq -s .\
      | tee modules_taken.json
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
