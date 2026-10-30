<?php

// This suite runs under phpunit on the php binary only. It is not linted or
// executed by phpscript, so PHP syntax phpscript cannot parse (keyed yield)
// is fine here; atkins.yml lints tests by fixture name for the same reason.

declare(strict_types=1);

namespace MiniTPL\Tests;

use MiniTPL\Compiler;
use MiniTPL\Template;
use PHPUnit\Framework\Attributes\CoversClass;
use PHPUnit\Framework\Attributes\DataProvider;
use PHPUnit\Framework\TestCase;

#[CoversClass(Template::class)]
#[CoversClass(Compiler::class)]
final class TemplateTest extends TestCase
{
	private const TEMPLATES = 'tests/templates/';
	private const TEMPLATES_BROKEN = 'tests/templates2/';
	private const COMPILE = 'tests/compile/';
	private const EXPECTED = 'tests/compiled/';

	/** Compiled output produced by {@see Template::load()} lands next to the templates. */
	private const RUNTIME_COMPILE = self::TEMPLATES . 'tests/';

	protected function setUp(): void
	{
		self::rmdir(self::RUNTIME_COMPILE);
	}

	protected function tearDown(): void
	{
		self::rmdir(self::RUNTIME_COMPILE);
	}

	#[DataProvider('compileProvider')]
	public function testCompile(string $template): void
	{
		$tpl = $this->template();

		$destination = self::COMPILE . $template;
		$result = $tpl->compile(self::TEMPLATES . $template, $destination);

		$this->assertSame(1, $result, 'compile() should report success');
		$this->assertFileEquals(self::EXPECTED . $template, $destination);
	}

	/** @return iterable<string, array{string}> */
	public static function compileProvider(): iterable
	{
		$templates = glob(__DIR__ . '/templates/*.tpl') ?: [];
		sort($templates);

		foreach ($templates as $template) {
			$name = basename($template);
			yield $name => [$name];
		}
	}

	public function testRenderMatchesGet(): void
	{
		$tpl = $this->template();
		$this->assertTrue($tpl->load('08_utf8_bom.tpl'));

		$this->assign($tpl);

		$returned = $tpl->get();

		ob_start();
		$tpl->render();
		$echoed = ob_get_clean();

		$this->assertNotSame('', $returned);
		$this->assertSame($returned, $echoed);
	}

	public function testStaleCompiledTemplateIsRecompiled(): void
	{
		$tpl = $this->template();
		$this->assertTrue($tpl->load('08_utf8_bom.tpl'));

		$compiled = self::RUNTIME_COMPILE . 'compile/08_utf8_bom.tpl';
		$this->assertFileExists($compiled);

		// Backdate the compiled file so the next load() has to recompile it.
		touch($compiled, filemtime(self::TEMPLATES . '08_utf8_bom.tpl') - 86400);
		$stale = filemtime($compiled);

		$this->assertTrue($tpl->load('08_utf8_bom.tpl'));
		clearstatcache(true, $compiled);
		$this->assertGreaterThan($stale, filemtime($compiled));

		$this->assign($tpl);
		$this->assertNotSame('', $tpl->get());
	}

	/**
	 * A compile is never visible half-written.
	 *
	 * The compiled template is executed with include(), so a reader that
	 * catches it mid-write executes a truncated program - a page that renders
	 * its doctype and its head and then stops. That is what opening the target
	 * with mode "w" allowed: the truncate lands before the first byte of the
	 * new content.
	 *
	 * Two things are asserted. The target is never zero bytes after a compile,
	 * and no temporary is left behind - a rename that failed would leave one,
	 * and a caller would then be reading a stale file forever without knowing.
	 */
	public function testCompileIsNotVisibleHalfWritten(): void
	{
		$tpl = $this->template();
		$this->assertTrue($tpl->load('08_utf8_bom.tpl'));

		$compiled = self::RUNTIME_COMPILE . 'compile/08_utf8_bom.tpl';
		$this->assertFileExists($compiled);
		$this->assertGreaterThan(0, filesize($compiled));

		$before = file_get_contents($compiled);

		// Recompile over a file that already holds a complete program. Under
		// the old write this truncated first; under the rename the previous
		// content stays readable until the new one is complete.
		touch($compiled, filemtime(self::TEMPLATES . '08_utf8_bom.tpl') - 86400);
		$this->assertTrue($tpl->load('08_utf8_bom.tpl'));

		clearstatcache(true, $compiled);
		$this->assertGreaterThan(0, filesize($compiled), 'the compiled template was left empty');
		$this->assertSame($before, file_get_contents($compiled), 'a recompile of an unchanged template changed its bytes');

		$leftovers = glob(self::RUNTIME_COMPILE . 'compile/*.tmp');
		$this->assertSame([], $leftovers, 'a temporary compile file was left behind');
	}

	public function testFindPathReturnsFalseForUnknownTemplate(): void
	{
		$this->assertFalse($this->template()->_find_path('404.tpl'));
	}

	#[DataProvider('compileLocationProvider')]
	public function testCompilePath(string $location, bool $absolute, string $expected): void
	{
		$tpl = $this->template();
		$tpl->set_compile_location($location, $absolute);

		$this->assertSame($expected, $tpl->_compile_path(self::TEMPLATES));
	}

	/** @return iterable<string, array{string, bool, string}> */
	public static function compileLocationProvider(): iterable
	{
		yield 'relative' => ['tests/compile/', false, 'tests/templates/tests/compile/'];
		yield 'absolute' => ['/tests/compile/', true, '/tests/compile/tests/templates/'];
	}

	/**
	 * Without a call to set_compile_location(), the compiled templates are
	 * stored beside the source templates, in a "<dir>_c" directory.
	 */
	public function testCompileLocationDefaultsToAdjacentTemplatesC(): void
	{
		$tpl = new Template();

		$this->assertSame('templates_c/', $tpl->_compile_path('templates/'));
		$this->assertSame('theme/templates_c/', $tpl->_compile_path('theme/templates/'));
		$this->assertSame('templates_c/', $tpl->_compile_path('./'));
	}

	public function testGetVar(): void
	{
		$tpl = new Template();
		$tpl->assign('foo', 'bar');

		$this->assertSame('bar', $tpl->getVar('foo'));
		$this->assertFalse($tpl->getVar('missing'));
	}

	public function testLoadThrowsWhenTemplateCannotBeCompiled(): void
	{
		$tpl = new Template();
		$tpl->set_compile_location(self::COMPILE, false);
		$tpl->set_paths(self::TEMPLATES_BROKEN);

		$this->expectException(\Exception::class);
		$this->expectExceptionMessage("Template file 'fail_to_compile.tpl' doesn't exist!");

		$tpl->load('fail_to_compile.tpl');
	}

	public function testRenderThrowsWhenNothingWasLoaded(): void
	{
		$tpl = new Template();

		$this->assertFalse($tpl->load('missing.tpl'));

		$this->expectException(\Exception::class);
		$this->expectExceptionMessage("Filename can't be empty, tried to render 'missing.tpl'");

		$tpl->render();
	}

	#[DataProvider('varsProvider')]
	public function testSplitExpression(string $expression, string $expected, string $description): void
	{
		$compiler = new Compiler();

		$this->assertSame($expected, $compiler->_split_exp($expression), $description);
	}

	public function testSplitExpressionRecognizesPhpTokenConstants(): void
	{
		$compiler = new Compiler();

		$this->assertSame("\$_v['variable']", $compiler->_split_exp('$variable'));
		$this->assertSame("\$_v['object']->method()", $compiler->_split_exp('$object->method()'));
	}

	/** @return iterable<string, array{string, string, string}> */
	public static function varsProvider(): iterable
	{
		yield 'normal string' => ['news_section_news_list.tpl', 'news_section_news_list.tpl', 'normal string'];
		yield 'variable' => ['$var', "\$_v['var']", 'variable'];
		yield 'array index' => ['$var.netko', "\$_v['var']['netko']", 'array index'];
		yield 'string concat' => ['$var . "netko"', "\$_v['var'] . \"netko\"", 'string concat'];
		yield 'variable concat' => ['$var1 . $var2', "\$_v['var1'] . \$_v['var2']", 'variable concat'];
		yield 'array var index' => ['$var1.$var2', "\$_v['var1'][\$_v['var2']]", 'array var index'];
		yield 'array int index' => ['$items.0', "\$_v['items']['0']", 'array int index'];
		yield 'object function' => ['$tpl->get()', "\$_v['tpl']->get()", 'object function'];
		yield 'second object function' => ['$tplx->get()', "\$_v['tplx']->get()", 'second object function'];
	}

	#[DataProvider('contextProvider')]
	public function testHtmlContext(string $document, string $needle, string $expected): void
	{
		$compiler = new Compiler();

		$offset = strpos($document, $needle);
		$this->assertNotFalse($offset, 'the needle has to occur in the document');

		$map = $compiler->_html_map($document);

		$this->assertSame($expected, $compiler->_context_at($map, $offset));
	}

	/** @return iterable<string, array{string, string, string}> */
	public static function contextProvider(): iterable
	{
		yield 'text' => ['<p>{v}</p>', '{v}', Compiler::CONTEXT_TEXT];
		yield 'bare attribute' => ['<p title={v}>', '{v}', Compiler::CONTEXT_TAG];
		yield 'double quoted attribute' => ['<p title="{v}">', '{v}', Compiler::CONTEXT_ATTRIBUTE_DQ];
		yield 'single quoted attribute' => ["<p title='{v}'>", '{v}', Compiler::CONTEXT_ATTRIBUTE_SQ];
		yield 'comment' => ['<!-- {v} -->', '{v}', Compiler::CONTEXT_COMMENT];
		yield 'script body' => ['<script>{v}</script>', '{v}', Compiler::CONTEXT_RAW];
		yield 'style body' => ['<style>{v}</style>', '{v}', Compiler::CONTEXT_RAW];
		yield 'uppercase script body' => ['<SCRIPT>{v}</SCRIPT>', '{v}', Compiler::CONTEXT_RAW];
		yield 'after script body' => ['<script>x</script><p>{v}</p>', '{v}', Compiler::CONTEXT_TEXT];
		yield 'after comment' => ['<!-- x --><p>{v}</p>', '{v}', Compiler::CONTEXT_TEXT];
		yield 'generated php' => ['<?php echo "{v}"; ?>', '{v}', Compiler::CONTEXT_PHP];
		yield 'after generated php' => ['<?php echo 1; ?><p>{v}</p>', '{v}', Compiler::CONTEXT_TEXT];
		yield 'php inside attribute' => ['<p title="<?php echo 1; ?>{v}">', '{v}', Compiler::CONTEXT_ATTRIBUTE_DQ];
		yield 'closing tag is not raw' => ['<script>x</script>{v}', '{v}', Compiler::CONTEXT_TEXT];
		yield 'unterminated tag' => ['<p title="a', 'title', Compiler::CONTEXT_TAG];
		yield 'unterminated attribute' => ['<p title="{v}', '{v}', Compiler::CONTEXT_ATTRIBUTE_DQ];
		yield 'unterminated comment' => ['<!-- {v}', '{v}', Compiler::CONTEXT_COMMENT];
		yield 'unterminated script' => ['<script>{v}', '{v}', Compiler::CONTEXT_RAW];
		yield 'unterminated php' => ['<?php echo "{v}";', '{v}', Compiler::CONTEXT_PHP];
		yield 'text with no markup' => ['plain {v} text', '{v}', Compiler::CONTEXT_TEXT];
	}

	#[DataProvider('escapeProvider')]
	public function testRenderEscapesByContext(string $source, array $vars, string $expected): void
	{
		$this->assertSame($expected, $this->render($source, $vars));
	}

	/** @return iterable<string, array{string, array<string, string>, string}> */
	public static function escapeProvider(): iterable
	{
		$hostile = ['v' => '<b>"x" & \'y\'</b>'];
		$escaped = '&lt;b&gt;&quot;x&quot; &amp; &#039;y&#039;&lt;/b&gt;';

		yield 'text node' => ['<p>{v}</p>', $hostile, '<p>' . $escaped . '</p>'];
		yield 'attribute' => ['<a href="{v}">x</a>', $hostile, '<a href="' . $escaped . '">x</a>'];
		yield 'single quoted attribute' => ["<a href='{v}'>x</a>", $hostile, "<a href='" . $escaped . "'>x</a>"];
		yield 'explicit escape' => ['{v|escape}', $hostile, $escaped];
		yield 'raw' => ['{v|raw}', $hostile, $hostile['v']];
		yield 'unescape alias' => ['{v|unescape}', $hostile, $hostile['v']];
		yield 'script body' => ['<script>{v}</script>', $hostile, '<script>' . $hostile['v'] . '</script>'];
		yield 'modifier is escaped' => ['{v|tolower}', $hostile, '&lt;b&gt;&quot;x&quot; &amp; &#039;y&#039;&lt;/b&gt;'];
		yield 'modifier chain with raw' => ['{v|toupper|raw}', ['v' => '<b>x</b>'], '<B>X</B>'];
		yield 'same variable in two contexts' => [
			'<a title="{v}">{v|raw}</a>',
			['v' => '<i>'],
			'<a title="&lt;i&gt;"><i></a>',
		];
		yield 'noescape directive' => ['{*noescape*}<p>{v}</p>', $hostile, '<p>' . $hostile['v'] . '</p>'];
	}

	public function testConstantsEscapeByContext(): void
	{
		define('_TEST_SITE_NAME', 'Tom & "Jerry"');

		$this->assertSame(
			'<p title="Tom &amp; &quot;Jerry&quot;">Tom &amp; &quot;Jerry&quot;</p>',
			$this->render('<p title="{_TEST_SITE_NAME}">{_TEST_SITE_NAME}</p>', []),
			'a constant holds whatever define() was given it, so it escapes like any other value'
		);
		$this->assertSame(
			'Tom & "Jerry"',
			$this->render('{_TEST_SITE_NAME|raw}', []),
			'raw opts a constant out the same way it does a variable'
		);
		$this->assertSame(
			'<script>Tom & "Jerry"</script>',
			$this->render('<script>{_TEST_SITE_NAME}</script>', []),
			'a script body is not markup, so a constant is left alone there'
		);
	}

	public function testConstantsStillExpandInsideLiteralBlocks(): void
	{
		// Constants are replaced before <script type="text/template"> bodies
		// are stashed, which is what lets them work where variables do not.
		$compiled = $this->compileSource(
			'<script type="text/template">{nuts} and {_TEST_LITERAL_CONST}</script>'
		);

		$this->assertStringContainsString('{nuts}', $compiled, 'variables stay literal');
		$this->assertStringContainsString('_TEST_LITERAL_CONST', $compiled);
		$this->assertStringNotContainsString('{_TEST_LITERAL_CONST}', $compiled);
	}

	/** Compile a template source and return the generated php. */
	private function compileSource(string $source): string
	{
		$name = 'compile_' . md5($source) . '.tpl';
		file_put_contents(self::TEMPLATES . $name, $source);

		try {
			$destination = self::COMPILE . $name;
			$this->assertSame(1, $this->template()->compile(self::TEMPLATES . $name, $destination));

			return (string) file_get_contents($destination);
		} finally {
			unlink(self::TEMPLATES . $name);
		}
	}

	public function testSetEscapeDisablesAutoEscaping(): void
	{
		$this->assertSame('<p><i></p>', $this->render('<p>{v}</p>', ['v' => '<i>'], false));
	}

	/** Compile and render a template source, returning its output. */
	private function render(string $source, array $vars, bool $escape = true): string
	{
		$name = 'render_' . md5($source . var_export($escape, true)) . '.tpl';
		file_put_contents(self::TEMPLATES . $name, $source);

		try {
			$tpl = $this->template();
			$tpl->set_escape($escape);
			$this->assertTrue($tpl->load($name));

			foreach ($vars as $key => $value) {
				$tpl->assign($key, $value);
			}

			return $tpl->get();
		} finally {
			unlink(self::TEMPLATES . $name);
		}
	}

	/** A template configured against the test fixtures. */
	private function template(): Template
	{
		$tpl = new Template();
		$tpl->set_paths(self::TEMPLATES);
		$tpl->set_compile_location(self::COMPILE, false);
		$tpl->add_default('key', 'val');

		return $tpl;
	}

	/** The variables the fixture templates expect. */
	private function assign(Template $tpl): void
	{
		$tpl->assign('items', [['id' => 1], ['id' => 2], ['id' => 3]]);
		$tpl->assign(['foo' => 'bar', 'd' => ['burger']], 'foo');
		$tpl->assign('.foo_foo', 'baz');
		$tpl->assign('.foo_d', ['steak', 'beef', 'pork', 'chicken']);
	}

	private static function rmdir(string $path): void
	{
		if (!is_dir($path)) {
			return;
		}

		foreach (scandir($path) ?: [] as $entry) {
			if ($entry === '.' || $entry === '..') {
				continue;
			}
			$child = $path . '/' . $entry;
			is_dir($child) ? self::rmdir($child) : unlink($child);
		}

		rmdir($path);
	}
}
