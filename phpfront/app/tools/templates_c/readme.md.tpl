<?php $_v=&$this->vars;?>
# <?php echo $_v['title'];?>


Generated from the module by `atkins docs`. DO NOT EDIT.

<?php if($_v['tables']){?>
![The tables this module reads](schema.svg)

<?php }?><?php echo $_v['description'];?>


## Pages

<?php if(!$_v['routes']){?>
This module draws no page. It owns a data model other modules read, so it
carries a schema and the statements over it and no route.

<?php }else{?>
| Method | Path | Answered by |
|--------|------|-------------|
<?php if(!empty($_v['routes']))foreach($_v['routes'] as $_v['route']){?>| `<?php echo $_v['route']['method'];?>
` | `<?php echo $_v['route']['path'];?>
` | [`<?php echo $_v['route']['file'];?>
`](<?php echo $_v['route']['file'];?>
) |<?php echo $_v['nl']; } }?>## The data model

<?php if(!$_v['tables']){?>
No page of this module reads a table of its own. The chrome every page
draws reads the instance's, which `docs/phpfront.md` lists once.

<?php }?><?php if(!empty($_v['tables']))foreach($_v['tables'] as $_v['table']){?>### `<?php echo $_v['table']['name'];?>
`

<?php echo $_v['table']['comment'];?>


| Column | Type | Key |
|--------|------|-----|
<?php if(!empty($_v['table']['columns']))foreach($_v['table']['columns'] as $_v['column']){?>| `<?php echo $_v['column']['name'];?>
` | `<?php echo $_v['column']['type'];?>
` | <?php echo $_v['column']['key'];?>
 |<?php echo $_v['nl']; } }?>## Operations

<?php if(!empty($_v['classes']))foreach($_v['classes'] as $_v['class']){?>### `<?php echo $_v['class']['name'];?>
`

<?php echo $_v['class']['doc'];?>


| Method | What it does |
|--------|--------------|
<?php if(!empty($_v['class']['methods']))foreach($_v['class']['methods'] as $_v['method']){?>| `<?php echo $_v['method']['signature'];?>
` | <?php echo $_v['method']['doc'];?>
 |<?php echo $_v['nl']; } }?><?php if(!$_v['classes']){?>
This module owns no data model of its own. It reads the instance's tables
through `code/Task`.

<?php }?>---

The tables come from `tables()` on <?php echo $_v['module'];?>
. The operations are the classes of
`storage/`, which hold every statement this module runs.
