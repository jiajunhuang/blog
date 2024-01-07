# OpenAI Prompt Engineering 摘录和总结

这篇文章是指导如何写出比较好的prompt，本文是我的摘录和总结。

原文链接：https://platform.openai.com/docs/guides/prompt-engineering/strategy-write-clear-instructions

策略:

- 写出清晰的指令
    - 写清楚，写详细。比如"Who’s president?"就不清晰，OpenAI就得猜测你的意思，写成 "Who was the president of Mexico in 2021, and how frequently are elections held?" 就清晰很多
    - 明确告诉OpenAI它要扮演的角色。例如，生成SQL时，写上"You're a MySQL expert"，生成的SQL就会更符合MySQL标准，质量也更高
    - 用户的输入，要明确的分隔符包裹起来，例如：

    ```
    Summarize the text delimited by triple quotes with a haiku.

    """insert text here"""
    ```

    - 明确告诉OpenAI，完成任务所需要的步骤（`step-by-step` 很重要），例如：

    ```
Use the following step-by-step instructions to respond to user inputs.

Step 1 - The user will provide you with text in triple quotes. Summarize this text in one sentence with a prefix that says "Summary: ".

Step 2 - Translate the summary from Step 1 into Spanish, with a prefix that says "Translation: ".
    ```

    - 如果希望用户按照你的模板输出，可以给OpenAI举一个例子
    - 可以明确告诉OpenAI希望的输出长度。OpenAI很喜欢废话文学，如果你希望简练一些，或者讲的更详细一些，那么你就得明确告诉它。

- 给出参考链接
    - 当你希望OpenAI从你给定的资料里查找答案时，需要明确给出资料
    - OpenAI 遇到不会的问题的时候，它会开始瞎编而且编的有模有样。避免这种现象，可以要求OpenAI列出引用来源。

- 把复杂的问题拆分成简单的子任务
    - 比如对于长文章的总结，可以每章总结，然后再往上一层再次总结

- 给模型时间“思考”
    - 一些简单的问题，OpenAI为了节省时间，可能会乱说，比如问他某个问题是不是正确的，OpenAI可能会直接说问题是正确的。可以加上一些指示，要求它必须确认没问题，才进行回答："Don't decide if the student's solution is correct until you have done the problem yourself."
    - 明确告诉OpenAI，质量很重要，可以慢慢回答，这个是我在实践中总结的一条经验，效果还是很不错的

- 使用外部工具
    - 比如embeddings-based search，例如，如果用户询问有关特定电影的问题，将有关电影的高质量信息（例如演员、导演等）添加到模型的输入中可能会有用。嵌入可用于实现高效的知识检索，以便在运行时动态地将相关信息添加到模型输入中。
    - 比如给一段代码，要求OpenAI执行

- 逐步改进prompt，系统的测试。日常实践中，构建的AI应用，一套评估测试体系是很有必要的，这样可以及时反馈出AI在生成结果中的变化
    - 改善prompt是一个循环：先尝试写出第一版清晰具体的prompt，测试并分析结果，改进prompt，继续分析测试不断循环不断迭代

---

ref:

这里也有一个吴恩达和OpenAI联合推出的视频教程，涵盖了上面的内容，还包括一些其他的主题例如翻译、文本扩写等：

- https://space.bilibili.com/15467823/channel/collectiondetail?sid=1354861
