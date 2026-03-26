import { defineBuildConfig } from 'unbuild';

export default defineBuildConfig({
  clean: true,
  declaration: false,
  entries: [
    {
      builder: 'mkdist',
      input: './src/scss-bem',
      pattern: ['**/*.scss'],
    },
    {
      builder: 'mkdist',
      input: './src/css',
      pattern: ['**/*.css'],
    },
    'src/index',
  ],
});
