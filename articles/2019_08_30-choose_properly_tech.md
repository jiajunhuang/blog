# 选择合适的技术栈

刚毕业的新人特别喜欢使用新的技术，从业多年的老人往往偏于保守。

```
- k8s好用，我们快上吧！
- supervisor用着不挺好的吗？
- 这都9102年了，还不上docker？

- 什么，你们还在用jQuery？
- 为什么不用jQuery？
```

选择合适的技术栈，这大概是对程序员非常重要的一个考验，k8s固然听起来很好用，但是对于小团队来说维护成本相当高，而它
所标榜的快速扩容，在云服务商管理台点一下复制也是完全没有问题的，k8s标榜的自动扩容实际生产中根本没有几个人敢用。

- 对于新的技术，我们要持保守态，它真的解决了问题吗？解决了我们的痛点吗？现有技术栈能否做到同样的效果呢？要知道，我们的
规模、技术人员水平和Google都不是一个层次的。脱离实际业务选技术栈，就是耍流氓。

再举个例子，如果是做一个扫码点餐系统，对于客户侧，你会选择什么技术栈呢？无脑上React/Vue/AngularJS吗？不对，实际上的
业务场景并非如此，用户都是扫完就走的，所以对于页面加载速度有非常高的要求，而React等现代前端框架，动辄几百k，只要网速
稍差，加载时间都会受影响，且扫码页面并不复杂，传统的模板渲染会是更好的选择。

- 对于老的技术，我们要评估是否该升级。相比传统的裸机部署，Docker的确提升了很大的部署体验，解决了依赖问题，极大的解决了
各个机器环境不一致的痛点，对于这种技术，我们应该果断的上。

如何选择合适的技术栈？结合业务场景和业务需求，不要偏激，也不要固守陈规，不管白猫黑猫，抓到老鼠的就是好猫。

---

参考资料：

- [k8s并不适合小工程](http://carlosrdrz.es/kubernetes-for-small-projects/)
- [选择无聊的技术](https://boringtechnology.club/)