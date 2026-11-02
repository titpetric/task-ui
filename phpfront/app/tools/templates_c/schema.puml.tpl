<?php $_v=&$this->vars;?>
@startuml
' Generated from schema/schema.yml by tools/docs.php. DO NOT EDIT.
hide circle
hide empty members
skinparam linetype ortho
skinparam shadowing false
skinparam backgroundColor transparent
<?php if(!empty($_v['tables']))foreach($_v['tables'] as $_v['table']){?>entity "<?php echo $_v['table']['name'];?>
" {<?php echo $_v['nl']; if(!empty($_v['table']['columns']))foreach($_v['table']['columns'] as $_v['column']){?>  <?php echo $_v['column']['mark'];?>
 <?php echo $_v['column']['name'];?>
 : <?php echo $_v['column']['type']; echo $_v['nl']; }?>}<?php echo $_v['nl']; echo $_v['nl']; }?><?php if(!empty($_v['links']))foreach($_v['links'] as $_v['link']){?><?php echo $_v['link']['from'];?>
 }o--|| <?php echo $_v['link']['to'];?>
 : <?php echo $_v['link']['column']; echo $_v['nl']; }?>@enduml
