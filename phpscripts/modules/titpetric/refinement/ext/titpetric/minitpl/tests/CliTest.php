<?php

declare(strict_types=1);

namespace MiniTPL\Tests;

use PHPUnit\Framework\TestCase;

/**
 * Exercises cmd/minitpl as a process, the way a build or a person calls it.
 *
 * The command line runs under phpscript, which roots its source filesystem at
 * the directory it starts in and takes no path outside it. The fixture tree
 * therefore lives inside the checkout, below a directory this test creates,
 * and the command is called with paths relative to the project root.
 */
final class CliTest extends TestCase
{
	private string $root;
	private string $base;
	private string $cli;
	private string $phpscript;

	protected function setUp(): void
	{
		$phpscript = trim((string) shell_exec('command -v phpscript 2>/dev/null'));
		if ($phpscript === '') {
			$this->markTestSkipped('phpscript is not installed; the command line runs under it');
		}

		$this->phpscript = $phpscript;
		$this->cli = dirname(__DIR__) . '/cmd/minitpl';
		$this->base = '.cli-tmp-' . bin2hex(random_bytes(8));
		$this->root = dirname(__DIR__) . '/' . $this->base;

		$this->write('templates/page.tpl', "<h1>{title}</h1>\n{include footer.tpl}\n");
		$this->write('templates/footer.tpl', "<footer>{year}</footer>\n");
		$this->write('templates/partials/note.tpl', "<p>{note}</p>\n");
		$this->write('data.yml', "title: Hello\nyear: 2026\n");
	}

	protected function tearDown(): void
	{
		if (!isset($this->root)) {
			return;
		}

		self::rmdir($this->root);
	}

	public function testCompileWritesTemplatesC(): void
	{
		[$status, $stdout, $stderr] = $this->runCli(['compile', $this->arg('templates')]);

		$this->assertSame(0, $status, $stderr);
		$this->assertFileExists($this->host('templates_c/page.tpl'));
		$this->assertFileExists($this->host('templates_c/footer.tpl'));
		$this->assertFileExists($this->host('templates_c/partials/note.tpl'));

		$compiled = (string) file_get_contents($this->host('templates_c/page.tpl'));
		$this->assertStringContainsString("\$_v['title']", $compiled);
		$this->assertStringContainsString('<footer>', $compiled, 'the include is expanded at compile time');
		$this->assertStringContainsString('compiled 3 templates into ' . $this->arg('templates_c'), $stdout);
	}

	public function testCompileIntoExplicitOutputQuietly(): void
	{
		[$status, $stdout, $stderr] = $this->runCli([
			'compile', '--quiet', $this->arg('templates'), $this->arg('build'),
		]);

		$this->assertSame(0, $status, $stderr);
		$this->assertSame('', $stdout);
		$this->assertFileExists($this->host('build/page.tpl'));
		$this->assertDirectoryDoesNotExist($this->host('templates_c'));
	}

	public function testCompileReportsAnUncompilableTemplate(): void
	{
		$this->write('templates/empty.tpl', '');

		[$status, , $stderr] = $this->runCli(['compile', '-q', $this->arg('templates')]);

		$this->assertSame(1, $status);
		$this->assertStringContainsString("failed to compile '" . $this->arg('templates/empty.tpl') . "'", $stderr);
		$this->assertFileExists($this->host('templates_c/page.tpl'), 'the other templates still compile');
	}

	public function testCompileNeedsADirectory(): void
	{
		[$status, , $stderr] = $this->runCli(['compile', $this->arg('missing')]);

		$this->assertSame(1, $status);
		$this->assertStringContainsString("templates directory '" . $this->arg('missing') . "' does not exist", $stderr);
	}

	public function testRenderWithDataFile(): void
	{
		[$status, $stdout, $stderr] = $this->runCli([
			'render', $this->arg('templates/page.tpl'), $this->arg('data.yml'),
		]);

		$this->assertSame(0, $status, $stderr);
		$this->assertSame("<h1>Hello</h1>\n<footer>2026</footer>\n\n", $stdout);
		$this->assertFileExists(
			$this->host('templates_c/page.tpl'),
			'rendering compiles beside the template directory, without configuration'
		);
	}

	public function testRenderReadsDataFromStandardInput(): void
	{
		[$status, $stdout, $stderr] = $this->runCli(
			['render', $this->arg('templates/page.tpl')],
			"title: Piped\nyear: 1\n"
		);

		$this->assertSame(0, $status, $stderr);
		$this->assertSame("<h1>Piped</h1>\n<footer>1</footer>\n\n", $stdout);
	}

	public function testRenderReadsDashAsStandardInput(): void
	{
		[$status, $stdout, $stderr] = $this->runCli(
			['render', $this->arg('templates/page.tpl'), '-'],
			"title: Dash\nyear: 2\n"
		);

		$this->assertSame(0, $status, $stderr);
		$this->assertSame("<h1>Dash</h1>\n<footer>2</footer>\n\n", $stdout);
	}

	public function testRenderEscapesValuesInMarkup(): void
	{
		[$status, $stdout, $stderr] = $this->runCli(
			['render', $this->arg('templates/page.tpl')],
			"title: \"<b>x</b>\"\nyear: 3\n"
		);

		$this->assertSame(0, $status, $stderr);
		$this->assertStringContainsString('<h1>&lt;b&gt;x&lt;/b&gt;</h1>', $stdout);
	}

	public function testRenderNeedsAnExistingTemplate(): void
	{
		[$status, , $stderr] = $this->runCli(['render', $this->arg('missing.tpl')]);

		$this->assertSame(1, $status);
		$this->assertStringContainsString("template file '" . $this->arg('missing.tpl') . "' does not exist", $stderr);
	}

	public function testRenderRejectsInvalidYaml(): void
	{
		[$status, , $stderr] = $this->runCli(
			['render', $this->arg('templates/page.tpl')],
			"title: [1, 2\n"
		);

		$this->assertSame(1, $status);
		$this->assertStringContainsString('unable to parse template data', $stderr);
	}

	public function testRenderRejectsDataThatIsNotAMapping(): void
	{
		[$status, , $stderr] = $this->runCli(
			['render', $this->arg('templates/page.tpl')],
			"- one\n- two\n"
		);

		$this->assertSame(1, $status);
		$this->assertStringContainsString('template data has to be a yaml mapping', $stderr);
	}

	public function testRenderRejectsAMissingDataFile(): void
	{
		[$status, , $stderr] = $this->runCli([
			'render', $this->arg('templates/page.tpl'), $this->arg('missing.yml'),
		]);

		$this->assertSame(1, $status);
		$this->assertStringContainsString('Unable to read yaml file', $stderr);
	}

	public function testUnknownOptionIsAnError(): void
	{
		[$status, , $stderr] = $this->runCli(['render', '--bogus', $this->arg('templates/page.tpl')]);

		$this->assertSame(1, $status);
		$this->assertStringContainsString("unknown option '--bogus'", $stderr);
	}

	public function testHelp(): void
	{
		[$status, $stdout] = $this->runCli(['help']);

		$this->assertSame(0, $status);
		$this->assertStringContainsString('Usage:', $stdout);
		$this->assertStringContainsString('minitpl compile', $stdout);
	}

	public function testCompileHelpDoesNotCompile(): void
	{
		[$status, $stdout] = $this->runCli(['compile', '--help', $this->arg('templates')]);

		$this->assertSame(0, $status);
		$this->assertStringContainsString('Usage:', $stdout);
		$this->assertDirectoryDoesNotExist($this->host('templates_c'));
	}

	/**
	 * Run the command line with arguments and standard input.
	 *
	 * @param list<string> $args
	 * @return array{int, string, string} status, stdout, stderr
	 */
	private function runCli(array $args, string $stdin = ''): array
	{
		$command = array_merge([$this->phpscript, $this->cli], $args);
		$descriptors = [
			0 => ['pipe', 'r'],
			1 => ['pipe', 'w'],
			2 => ['pipe', 'w'],
		];

		$process = proc_open($command, $descriptors, $pipes, dirname(__DIR__));
		$this->assertIsResource($process);

		fwrite($pipes[0], $stdin);
		fclose($pipes[0]);

		$stdout = (string) stream_get_contents($pipes[1]);
		fclose($pipes[1]);

		$stderr = (string) stream_get_contents($pipes[2]);
		fclose($pipes[2]);

		return [proc_close($process), $stdout, $stderr];
	}

	/** The fixture path as the command line, running from the project root, names it */
	private function arg(string $relative): string
	{
		return $this->base . '/' . $relative;
	}

	/** The fixture path on the host, for the assertions */
	private function host(string $relative): string
	{
		return $this->root . '/' . $relative;
	}

	private function write(string $relative, string $contents): void
	{
		$path = $this->root . '/' . $relative;
		$dir = dirname($path);
		if (!is_dir($dir)) {
			mkdir($dir, 0777, true);
		}

		file_put_contents($path, $contents);
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
