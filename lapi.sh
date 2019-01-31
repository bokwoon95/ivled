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
  curl "https://ivle.nus.edu.sg/api/Lapi.svc/Workbins?APIKey=$LAPIkey&AuthToken=$AuthToken&CourseID=$1&TitleOnly=false&WorkbinID=$2" | jq .
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

modules_taken() {
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
}

export CG3002="417770fc-b119-488d-8c4b-bb70b88d7ac5"
export CS2102="8f2c5b07-a5be-4d86-bb1c-8839d5b72292"
export EE4204="fb80965b-191c-49c3-aa30-a5c890615721"
export EG2401="f559e6e3-eb62-45b4-97e3-cbdb57710eae"
export GES1002="1c752f6a-52f7-45f4-8a06-2d8b39878593"
export MA1512="7d6b57ce-4fd6-4f63-b94e-08c6306f985d"
export MA1513="08794bbd-b65a-4389-ad06-078b09fc729e"

workbins() {
  if [[ "$1" == "" ]]; then
    CourseID=$MA1513
  else
    CourseID=$1
  fi
  curl "https://ivle.nus.edu.sg/api/Lapi.svc/Workbins?APIKey=$LAPIkey&AuthToken=$AuthToken&CourseID=$CourseID"\
      | jq '.Results'
}

populate_jsons() {
  workbins $CG3002 > cg3002.json
  workbins $CS2102 > cs2102.json
  workbins $EE4204 > ee4204.json
  workbins $EG2401 > eg2401.json
  workbins $GES1002 > ges1002.json
  workbins $MA1512 > ma1512.json
  workbins $MA1513 > ma1513.json
}
