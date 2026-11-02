<?php $_v=&$this->vars;?>

<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<meta name="robots" content="noindex, nofollow">
	<title><?php if($_v['data']['title']){ echo htmlspecialchars($_v['data']['title'], ENT_QUOTES);?>
 | <?php }?>TaskUI</title>
	<link rel="icon" href="<?php echo htmlspecialchars($_v['data']['assets'], ENT_QUOTES);?>
/theme/icon.svg" type="image/svg+xml">
	<link rel="stylesheet" href="<?php echo htmlspecialchars($_v['data']['assets'], ENT_QUOTES);?>
/theme/vendor/xterm.css?v=<?php echo htmlspecialchars($_v['data']['assetVersion'], ENT_QUOTES);?>
">
	<link rel="stylesheet" href="<?php echo htmlspecialchars($_v['data']['assets'], ENT_QUOTES);?>
/theme/task-ui.css?v=<?php echo htmlspecialchars($_v['data']['assetVersion'], ENT_QUOTES);?>
">
</head>
