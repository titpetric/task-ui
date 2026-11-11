{*noescape*}@startuml
' Generated from schema/schema.yml by tools/docs.php. DO NOT EDIT.
hide circle
hide empty members
skinparam linetype ortho
skinparam shadowing false
skinparam backgroundColor transparent
{foreach $tables as $table}entity "{table.name}" {{nl}{foreach $table.columns as $column}  {column.mark} {column.name} : {column.type}{nl}{/foreach}}{nl}{nl}{/foreach}{foreach $links as $link}{link.from} }o--|| {link.to} : {link.column}{nl}{/foreach}@enduml
