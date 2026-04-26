import MarkdownIt from 'markdown-it'
import markdownItKatexExternal from 'markdown-it-katex-external'
import markdownItEmoji from 'markdown-it-emoji'
import markdownItContainer from 'markdown-it-container'
import markdownItSup from 'markdown-it-sup'
import markdownItSub from 'markdown-it-sub'
import markdownItFootnote from 'markdown-it-footnote'
import markdownItAbbr from 'markdown-it-abbr'
import markdownItIns from 'markdown-it-ins'
import markdownItMark from 'markdown-it-mark'
import markdownItKatex from '@iktakahiro/markdown-it-katex'
import mermaidPlugin from '@agoose77/markdown-it-mermaid'

export default function markdownToHtml(content: any) {
  const md = new MarkdownIt({
    html: true
  })
    .use(markdownItKatexExternal)
    .use(markdownItEmoji)
    .use(markdownItContainer, 'hljs-center')
    .use(markdownItContainer, 'hljs-left')
    .use(markdownItContainer, 'hljs-right')
    .use(markdownItSup)
    .use(markdownItSub)
    .use(markdownItFootnote)
    .use(markdownItAbbr)
    .use(markdownItIns)
    .use(markdownItMark)
    .use(markdownItKatex)
    .use(mermaidPlugin)
  return md.render(content)
}
