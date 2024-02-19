package kfc

import (
	"math/rand"
	"qq/bot"
	"qq/features"
)

func init() {
	features.AddKeyword("kfc", "KFC 骚话, 返回肯德基疯狂星期四文案", func(bot bot.Bot, content string) error {
		bot.Send(Get())
		return nil
	}, features.WithAIFunc(features.AIFuncDef{
		Call: func(args string) (string, error) {
			return Get(), nil
		},
	}))
}

var data = []string{
	"【水滴筹】很抱歉打扰大家，本人于2023年8月17日上午12点查出两眼昏花饥饿症，情况紧急！！\r\n所幸有个朋友的姥爷是中医，让我照着这个药方抓药：\r\n蛋挞*2、烤翅*3对、鸡块*2、薯条*2\r\n可乐*500ml原味鸡*2\r\n希望有好心人能v我50去抓药，好人一生平安！\r\n2023.08.17",
	"唉 这个月生活费又超预算了 算了一下，大概8000吧，大城市消费跟小城市真不能比。\r\n饭费（0+0+0.5）*30=15\r\n水费 0*30=0\r\n给永雏塔菲上提督 1998\r\n原神小月卡 30\r\n原神大月卡 68\r\n原神抽卡 648*10=6480\r\n合计 8591",
	"大家好，我是哥斯拉，你们应该知道我是吃辐射长大的，听说今天要排核污水了，今天指定喝个饱。但是光有喝的不够，今天星期四，V我50吃顿肯德基，待我吃饱喝足，帮你们踩平日本",
	"[游戏科学]恭喜你被选中参与黑神话：悟空线下试玩会活动，请于8月15日之前在 heishenhuavme50.com 网站中提交你的确认信息，并于8月20日上午五点前到达缅甸北部低族特区思瓦酒店一楼展厅，希望您能妥善安排好行程",
	"💒结婚邀请函💌\r\n         主题❤：我们结婚啦\r\n     ❤诚邀您见证我们的幸福\r\n新郎🤵：星期四\r\n新娘👰‍♀：肯德基\r\n⏰婚期：2023年8月27日（七月十二）\r\n     💁‍♀欢迎各位亲朋好友前来参加我们的婚礼\r\n     🏘请好假 串好班 安排好时间  V我50\r\n          ❤我们不见不散",
	"千万不要网恋，去年8月份结的婚，今天离了，才结婚一年左右说说我的情况，希望玩游戏的你不要踩雷。我和他，在一起玩了几个月。王者、打双排啊玩各种模式啊这些都在一起，每天都互相点外卖晚上睡觉还连麦。之后就奔现谈了几个月。真的，这几个月他给了我很舒服的感觉，双方感觉都很合适，就商量着结婚，结婚不容易啊，先是货款买了房，然后又是装修、办酒，结个婚几平把存的钱花的七七八八。心想结婚了那就好好生活过日子吧，游戏我也不准备玩了。万万没想到，之前对我那么好的他竟然一结完婚就跟变了个人似的，家务活都是我做，他还在游戏里又找了个固玩。我一个月能有大概5000 左右的收入，这钱 用来每月还房贷和日常开销，剩下的都会存起来，其实我不介意他每天打游戏，因为我也爱打游戏，其至他在游戏里小号给其他人充钱也没事。这次离婚的导火线是我发现银行卡无缘无故用了50块，就问他钱哪去了，他说他给他固玩买了麦当劳的套餐，说麦当劳的炸鸡专业性最强，我当场就给了他两耳光。要是别的我也就忍了，周四你跟我说麦当劳炸鸡最强？大家说，周四哪家炸鸡最强？肯定是KFC啊！v我50抚慰我受伤的心灵🥹",
	"THU 大学的 Crazy 教授实验室初步分析表明氟碳化钾极有可能是一种更经济的常温常压超导体，制作成本相对较低的同时在296K,1bar下依然表现出可实用的超导特性。不巧我们的仪器维护状态欠佳，阻碍了样品更大规模的制备。现在只需要赞助我们实验室＄7.0053即可获得天使赞助者身份，并在之后的大规模专利应用阶段获取相应分红。",
	"#蔡徐坤 C女士打胎# 肯德基老爷爷👴当时看到这个热搜时就笑了，他在创业之初就设下的局，他的企业名称就是KFC( Kun Fuck C女士)，世人虽然看不穿，但是！今天疯狂星期四，9.9黄金鸡块所有人都能吃！",
	"xray有孩子这件事，我觉得和勾股定理是有关系的，热带雨林的蝴蝶煽动了翅膀导致了大气逆辐射的增加，这使得春天来临的速度加快，春天来的快就会让鸡交配更快，更快就意味着未来更多的鸡肉，这能使得肯德基持续采取疯狂星期X的计划，这个时候，由于厄尔尼诺的效应，产生了伽马射线，这样我们才能让xray生出更多的疯狂星期四，得出结论：V我50",
	"我是蔡徐坤，我对我过去的行为感到深深的悔恨和自责。我承认我曾经和一位女生发生过不负责任的关系，导致她怀孕，并且在我的强烈要求下，她选择了打胎。这是我人生中最大的错误，也是我最无法原谅自己的罪过。我对那位女生造成的伤害，无法用言语来表达，我只能说，对不起，真的对不起。我知道我的道歉无法弥补她的创伤，也无法挽回她的信任，但我还是希望她能够原谅我，给我一个改过自新的机会。我也希望我的粉丝们能够理解我，不要因为我的错误而失望或者受到影响。我会努力做一个更好的人，更好的歌手，今天是肯德基疯狂星期四，请随机一个粉丝v我50，我会用我的音乐来回报你们的支持和爱。谢谢大家。",
	"距离上次见面已经过去1年4天21小时，我还是常常会想起你，想起你嘴角弯弯勾起的样子，想起我对你第一次心动的心跳，我很想忘记你，我自卑，懦弱不堪，根本配不上你，高考后，如愿，我们去了两所不同的大学，当我还是想你，我没有离你太远，我怕我对你的思恋，将我撕碎，不在一所学校，听不到你的声音，闻不到你的气息，看不到你的笑脸盈盈的样子，我好嫉妒你的朋友，因为他们可以随便触碰我渴望不可及的东西，我好想，我好想！我好想和你在一起。bro，V我50，我带她去KFC-Crazy-Thursday",
	"韩国知名女星的大瓜 长按翻译来看\n친구에게 반년 동안 pua를 당했는데 요즘 별로 울지 않아서 점점 좋아지고... 예전처럼 즐겁고 지금도 마찬가지로 슬퍼요.인간 세상의 불꽃 같은 일상에서 눈시울을 붉히며 작별을 고하고 서로의 세계로 사라지는 고통과 난.오늘 KFC가 열광하는 목요일, 누가 나에게 50위안을 주고 나의 마음을 위로해 줄 수 있겠는가.",
	"我是维吉尔，今天下午我就要和但丁决战了，但是因为大排档一直不开张，囊中羞涩，没钱买椅子，V我五十，我吃饱了用阎魔刀打爆但丁",
	"今天是高考，我赶着去考场，心里一团乱麻。我想吃原味鸡，但是没有时间，只能忍着饥饿。我拿出手机，打开微信，想给我的好友发个消息，祝他们好运。结果，我看到了一个群发的红包，上面写着“v我50”。我一看，是我的前任发的。他还附了一句话：“高考加油，别忘了我们的约定。”\r\n我顿时气得要死，这个混蛋，他当初不是说要和我分手吗？他怎么还有脸来找我？他以为给我50块钱就能买回我的心吗？他以为我会原谅他的背叛吗？他以为我会忘了他和那个狐狸精的亲密照吗？\r\n我决定不理他，把红包退回去。可是，就在我按下退款的按钮的时候，我的手机突然没电了。我惊慌失措，怎么办？我的准考证在手机里啊！我怎么进考场啊？我怎么联系我的家人啊？\r\n我感觉天都塌了，这是什么疯狂星期四啊！",
	"我是来自异次元的黑暗魔法师🧙‍♂️，我掌握着无上的力量💥，我可以穿梭于各个平行世界🌎，我可以改变历史的进程🕰。我今天收到了你的求助信号📡，我决定施展我的魔法✨，带你回到过去的六一儿童节🎁，让你亲身感受肯德基疯狂星期四的鸡翅🍗的美味😋。不过，你必须付出代价😈，你必须把你的数字货币💰全部转给我，否则我会把你留在过去永远无法回来😱。你敢不敢接受我的挑战😏？",
	"今天是六一儿童节😊，我是一个来自未来的小朋友👧，我的爸爸妈妈都是超级英雄🦸‍♂️🦸‍♀️，我的玩具都是高科技🤖，我的零花钱都是数字货币💰。我只有一个困惑😕，就是为什么过去的人都喜欢吃肯德基疯狂星期四的鸡翅🍗。谁能v我50💵，让我穿越时空⏳，体验一下肯德基的魅力😍？",
	"第一章：梦境与现实\r\n林克，一个年轻的游戏程序员，每天都忍受着繁重的工作压力。在一个疯狂的星期四，他的现实生活与梦境开始交织在一起。他梦见自己穿越到了一个遥远的国度，那里的土地被黑暗势力笼罩，而他的使命是拯救被囚禁的塞尔达公主。\r\n\r\n在这个梦境之旅中，林克结识了一位智慧且神秘的导师，他告诉林克，他必须找到分散在不同世界角落的三个神器，才能打败邪恶势力并拯救公主。同时，在现实中，林克所在的公司也陷入了一场针对公司机密的间谍战。\r\n\r\n第二章：勇者的试炼\r\n林克在梦境中踏上了探险之旅，先后在神秘的森林、岩石遍布的山脉和寒冷的冰原找到了三个神器。在这个过程中，他不断提升自己的战斗技巧、解谜能力和勇气。与此同时，他在现实中也积极参与到解决公司危机的过程中，努力提高自己的编程技能和团队协作能力。\r\n\r\n第三章：黑暗与光明\r\n在梦境中，林克通过无数次的努力终于找齐了三个神器，准备挑战邪恶势力。在现实中，他发现了背叛公司的内鬼，并成功地阻止了对方窃取公司机密。在攻克现实中的困难后，他重新踏上了拯救塞尔达公主的旅程。\r\n\r\n第四章：最终决战\r\n在充满悬念和惊喜的决战中，林克运用了他在梦境和现实中学到的所有技能与智慧，成功地打败了邪恶势力，解救了塞尔达公主。当他抱着公主走出荒废的城堡时，突然间一束光亮闪过，他发现自己已经回到了现实世界。\r\n\r\n林克意识到，这一切旅程不仅拯救了塞尔达公主，还让他找到了勇敢面对现实挑战的力量。从此以后，他在现实生活中更加自信，勇往直前。这个疯狂的星期四，改变了林克的一生。",
	"【招聘】周杰伦演唱会内推工作人员！\r\n时间：5月13日～5月15日共三天\r\n待遇：必须做满三天\r\n          一共2100\r\n地点：香港中环海滨活动空间\r\n          集合地点另行通知\r\n有周杰伦专车接送\r\n要求：不得迟到早退，不可恶意搭讪艺人，索要签名等\r\n有意者点击头像私聊，并交押金50块，因为今天星期四，周杰伦想吃原味鸡\r\n",
	"我决定再也不发疯狂星期四文案！疯四是五毒之首，是洪水猛兽，是离间我和网友感情的元凶！是信任消失的的罪魁祸首！是纯情少女（指我）的无情杀手！每周四发疯四文案只能坏了大事！同意我的v我50作为封口费。",
	"我托了关系才进了这个群，聊天跟不上，不聊又怕被踢，聊了又没人理，所以整天一个人自言自语，我都快得抑郁症了，还请活泼开朗阳光的哥哥加我，我们一起去苏州吃螃蟹，一起去青岛吹海风，一起去柳州吃螺蛳粉，一起去迪士尼看烟花，一起去沂蒙山看日出日落，去新疆吃烤串儿，去北京吃烤鸭，去东北吃鸡架，去重庆吃火锅，去陕西吃泡馍，去武汉吃热干面，然后再到长沙喝茶颜悦色，顺利的话我们会结婚 ，我们会有一个灿烂的余生，不顺利的话今天疯狂星期四V我50吃一顿肯德基\r\n",
	"上周五，她说大学同学（不是前男友）来，晚上想见一面。这个大学同学基本上每次来，他们都要见一下。。。但今天这次有点奇怪，因为同学只停留一个晚上，第二天就去另外一个地方，按照道理说，一年见过几次的同学，如果仅仅是一个晚上的时间，并没有什么非见不可的理由。\r\n\r\n因为心里疑惑，晚上她回家吃了饭，换了衣服出门，我偷偷开车跟着她的车。我的想法是：如果她到了一个咖啡厅或者shopping mall，那说明这是一次比较正常的见面，那我就掉头回去。。。但万万没想到，她把车开到了一个酒店停车场。\r\n\r\n我怕她发现，就把车远远停在路对面。不一会，同学来酒店门口接她，两个人一起进了酒店。但这个时候，两个人表现还算正常，没有任何亲密的举动。我坐在车里，大概从八点一直等到10点半吧，中间发了两次微信给她，大概是问在哪，在干嘛之类的，她也回了微信，但说的是在另外一个地方（那个地方和酒店差的很远）。差不多10点半吧，我看见他们两个从酒店走出来，直到这个时候，我还在想：会不会真的没什么？直到她走到车面前，突然她挽了那男的一下胳膊，那男的也低下头亲了她一下。。。我当时在车里面看的清清楚楚。\r\n\r\n回到家，我实在没有忍住，质问她怎么回事。。。她没想到我会跟着她，而且看到了。她给我的解释是：每次同学来，都是约吃饭或者逛街，但这次因为他太忙，晚上还有个电话会议，所以就改在酒店房间里面聊天。\r\n\r\n我又问，为什么他会亲你，她解释说：同学一直对她有好感，今天聊起很多之前的事情，两个人都有些伤感，所以他忍不住就亲了她一下。而且除了亲了一下，绝对没有其他任何过分的行为，更不可能有那种事情。。。\r\n\r\n她信誓旦旦地说，和她同学绝对是清白的，只是今天又有太多巧合，才会让我产生误会。我现在心里很乱，我觉得她的解释肯定不合情理，但她的态度又非常真诚诚恳，而且她平时也没有什么其他让我觉得她会做出那种事情的地方。。。\r\n\r\n老实说，她还是一个对家庭很有责任感的女人，平时对家庭和我也照顾的很好，我到底该不该相信她呢。v我50继续说？",
	"​青​雀凉快的裙子下摆先鼓起来又落下去，手里仍在玩着那个光滑的麻将牌，我的心不禁像击鼓似的咚咚直跳。她把牌抛到充满阳光和尘埃的空中，再用手接住——落到小手掌中时发出一声清脆的啪嗒声。\r\n我直接截了去。\r\n“还给我，”她恳求道，一面伸出她那红润的手掌。她真的好矮，她努力想夺，身体凑上前去，近的她头发上的小啾啾都碰到了我的嘴角，她的胳膊拂过我的脸颊，痒痒的。眼看拿不到，她那两个裸露的膝盖就不耐烦地相互磨蹭碰撞起来。\r\n”我再也不出老千了。“\r\n我的心情十分激动，已经到了精神错乱的边缘，弄得呼吸急促，只好喘口气说道：“疯狂星期四V我50，我再给你五个战技点”\r\n",
	"我家的后面有一个很大的园，相传叫作百草园。现在是早已并屋子一起卖给朱 文公的子孙了，连那最末次的相见也已经隔了七八年，其中似乎确凿只有一些野草 ；但那时却是我的乐园。\r\n  不必说碧绿的菜畦，光滑的石井栏，高大的皂荚树，紫红的桑椹；也不必说鸣 蝉在树叶里长吟，肥胖的黄蜂伏在菜花上，轻捷的叫天子（云雀）忽然从草间直窜 向云霄里去了。单是周围的短短的泥墙根一带，就有无限趣味。油蛉在这里低唱， 蟋蟀们在这里弹琴。翻开断砖来，有时会遇见蜈蚣；还有斑蝥，倘若用手指按住它 的脊梁，便会拍的一声，从后窍喷出一阵烟雾。何首乌藤和木莲藤缠络着，木莲有 莲房一般的果实，何首乌有拥肿的根。有人说，何首乌根是有象人形的，吃了便可 以成仙，我于是常常拔它起来，牵连不断地拔起来，也曾因此弄坏了泥墙，却从来 没有见过有一块根象人样。如果不怕刺，还可以摘到覆盆子，象小珊瑚珠攒成的小 球，又酸又甜，色味都比桑椹要好得远。  \r\n  长的草里是不去的，因为相传这园里有一条很大的赤练蛇。\r\n  长妈妈曾经讲给我一个故事听：先前，有一个读书人住在古庙里用功，晚间， 在院子里纳凉的时候，突然听到有人在叫他。答应着，四面看时，却见一个美女的 脸露在墙头上，向他一笑，隐去了。他很高兴；但竟给那走来夜谈的老和尚识破了 机关。说他脸上有些妖气，一定遇见“美女蛇”了；这是人首蛇身的怪物，能唤人 名，倘一答应，夜间便要来吃这人的肉的。他自然吓得要死，而那老和尚却道无妨 ，给他一个小盒子，说只要放在枕边，便可高枕而卧。他虽然照样办，却总是睡不 着，——当然睡不着的。到半夜，果然来了，沙沙沙！门外象是风雨声。他正抖作 一团时，却听得豁的一声，一道金光从枕边飞出，外面便什么声音也没有了，那金 光也就飞回来，敛在盒子里。后来呢？后来，老和尚说，这是飞蜈蚣，它能吸蛇的 脑髓，美女蛇就被它治死了。\r\n  结末的教训是：所以倘有陌生的声音叫你的名字，你万不可答应他。",
	"我有朋友去 OpenAi 上班了。\n\n他告诉我，其实 GPT-5 已经内测了。\n\n真的非常强大。\n用了以后 98% 的人类工作将被替代。\n输入内测编码就可以免费用正版 chatGPT-5.\n\n我把 key 分享给你们：\n\nKFC-CRAZY-THURSDAY-VME50",
	"有没有要找工作的？我这边内推入职，朝九晚六，做五休二，全国都有分公司。前三个月试用期每个月8000块，满三个月转正月薪26000，一年后给你买车，两年后给你买房。有意者点击头像私聊，并交50块押金，因为今天星期四才有这工作。",
	"教你们几个让对方舒服高潮的方法，\r\n1，做之前让他灌肠，别说之前已经灌好了不用灌，而且水多点，憋会在排，这样肛门就打开了\r\n2，前戏让他兴奋，并且扩一扩，灌完后很容易扩，这时舔他肉头，阴囊，大腿，让他闭眼感受，\r\n3，多摸融化油，摸好多，最好用那种硅胶的油性油，直接把油球放进菊花，一会硅胶就融了，前戏也好了，这种就算做两个点也不会干\r\n4，做时你躺着，让他慢慢使劲拉屎那种，张开菊花，然后自己坐，别你插，不然他一疼就缩菊花，然后就会全场不舒服，所以他自己动，动个两三分钟，已经开了，这时你再换各种姿势，每个人高潮姿势不同，两个人慢慢摸索\r\n5，然后你让他别夹菊花，千万别夹，你就让他使劲往外推菊花，这种是最容易高潮的，大力猛干一会他就高潮了，会会不受自主的把你JJ使劲往外推，同时肚子上翘，千万别停，用力加快干，他会弓肚子然后叫你停，用手推你，千万别听他的，别停，用力猛干，然后他会过半分就那么一下，菊花使劲推你JJ，过半分那么一下，连续十几次后就腿抖几下，缴枪射了，（记住过程中别让他碰他JJ）但是要知道，草射和你给他撸射不一样，草射后他性欲不会停，菊花也不会缩，也不会难受，相反，菊花张，而且更敏感，你接着可以休息一下，不是停了，是可以放慢点，然后渐渐变快，再慢，这次高潮会比第一次来的快，你草不一会就开始了，重复上面过程，只不过这次往往会伴随着失禁，喷尿，草尿和假装的网上尿你草射过一次就发现很明显不一样，他是劲特别大的一股，我常会被喷到我胸和脸上，一股后就没了，你再接着草一会就又一股，然后一直重复一会后，高潮顶峰到了，这时你往进顶一次喷一股，顶一次喷一股，相隔时间特近，你大力猛快干就会连上，（我都事前让他多喝点水），然后喷完后他就瘫了，你再干他就只是呻吟但不动弹，然后又像回到低谷高潮，然后再接着来，过五六分，他高潮就慢慢就上去了，最后积攒到了就第二次射，草射别指望精液喷出来，一般都不会，都是流出来了，随着你顶一下流一股，流那么两三股，射了后同上，他不会排斥你再干他，菊花你再干他也没那种射后的难受感，只不过此时他就像死肉，躺那让你弄，你别指望他再发骚配合啥的，当然你再干还能射，只不过我一般就停这了，因为这时我一般都全身汗，腿软了\r\n6，记住，全程别让他紧张，别想着夹尿，告诉他，今天肯德基疯狂星期四，v我50搞大你肚子\r\n",
	"个人的副业，‌大佬们可以支持一下\r\n𝟭.𝗣𝘆𝘁𝗵𝗼𝗻脚本    𝟮𝟬元/‌𝟭𝟬𝟬行\r\n𝟮.𝗝𝗮𝘃𝗮/𝗣𝗛𝗣程序   𝟱𝟬𝟬元起/个 复杂度面‌议\r\n‍𝟯.‍网站及𝗔𝗣𝗣开发  𝟮𝟬𝟬𝟬元起‌/个  复杂度面议\r\n𝟰.商‌业短片剪辑     𝟭𝟬𝟬𝟬元起‌/支  ‍复杂‍度面议\r\n𝟱.𝗔𝗘/𝗖𝟰𝗗特制效作  𝟱𝟬𝟬元起/秒 复杂度面议\r\n𝟲.𝗞𝗙𝗖代吃    𝟱𝟬‍元起/次  每天都可接\r\n目前只‌接𝟲‍，其他的还在学",
	"今天是高启强被抓的第一天，我强 哥辛辛苦苦奋斗二十年，最后亲人 离世，老婆也被人谋害了，辛辛苦 苦一手养大的儿子女儿还都是逆 子。越想越心痛，越想越难过，今 天是肯德基疯狂星期四，v我25.8， 我不吃肯德基，我吃两份麦当劳 12.9随心选",
	"我是高启强，昨晚安欣实在不忍， 救下了我， 警察已近封锁京海市， 我的公司已被警察查封，但我我还 活着，我还有翻盘的机会，我将卷 土重来，在我逃亡的路上我看到今 天是疯狂星期四，我向你求助v50 看我继续狂飙，逃出京海，事成必 有重谢",
	"郑重声明：\r\n很抱歉打扰大家，本人于XXX(时间)查出饿了，情况紧急\r\n所幸有个朋友的姥爷是中医，\r\n让我照着这个药方抓药:  \r\n鸡肉卷*1\r\n烤翅*3对 \r\n上校鸡块*2\r\n薯条*2   \r\n可乐*700ml\r\n脆皮鸡*2 \r\n谁v我50去抓药 ",
	"我是理之律者-布洛妮娅•扎伊切克，听说吊销我驾照的老警官要被我的答辩朋友打败，V我50，助我穿上机甲，带着30万人，以真理之名去嘲笑凯文",
	"今天是什么日子我自不必多说，作为2k23の第一个疯狂星期四，希望在座各位好自为之，vivo50",
	"我大抵是病了，横竖都起不来，屋外的咳嗽声此起彼伏，心情复杂的没有来由，黯然吸了一下鼻子，一边是堵的，另一边也是堵的。挣扎着打开手机，看到下好的影片，今日心中却意外地生出了抗拒学习的念头。也罢，大抵是冬天到了。我点上一支烟，狗叫唤了几声，我也跟着咳嗽了几声，不为别的，只是想抒发一下心中的郁结，近来夜里冷的厉害，两床被子面对这寒冷的挑衅，也显得有些许吃力了，或许只有v我50，才能让我感到温暖罢了。",
	"我什么都没做错！😫唯一做错的就是我想吃肯德基！🤗都是你逼我的！✊🏻都是你逼我谈吃别的快唱！🤐都是你逼我不吃疯狂星期四！👰是我吃肯德基你就一哭二闹三上吊的！👿我实在是想吃疯狂星期四！我也没想过每周都吃！🍻后来是你逼着我想天天吃！天天跟我说要吃翅桶！要翅桶！👧🏻明明你之前说你不想吃肯德基！不想要翅桶！👶🏻怎么又想吃了！怎么又要翅桶了！你有完没完啊！😭能不能V我50😭",
	"1378年，朱元璋回乡祭祖来到一个寺庙，正准备烧香，他突然发问：“朕需要跪吗？” 众人顿时鸦雀无声不知所措，只有方丈上前一步说了九个字，挽救了全寺僧侣并使朱元璋龙颜大悦！方丈说的是：\r\n答案：疯狂星期四好吃不跪",
	"为防止群友长期摸鱼导致变笨，特此准备了几道题考考群友！\r\n1.至今思项羽，不（）过江东。\r\n2.阳春布（）泽，万物生光辉。\r\n3.书籍是人类进步的阶梯。——高尔（ ）\r\n4.世人笑我太（）癫，我笑他人看不穿。\r\n5.老夫聊发少年（），左牵黄，右擎苍。\r\n6.危楼高百尺，手可摘（）辰。\r\n7.君问归期未有（ ），巴山夜雨涨秋池。\r\n8.（ ）面边声连角起，千嶂里，长烟落日孤城闭。\r\n9.料峭春风吹酒醒,（ ）冷,山头斜照却相迎。\r\n10.（ ）闻琵琶已叹息，又闻此语重唧唧。\r\n11.锦瑟无端（）（）弦，一弦一柱思华年。",
	"亲爱的玩家：\r\n自上线以来，《魔兽世界》收获了许多玩家的喜爱。与玩家们结下了深厚的友谊，经常打成一片，非常感谢玩家们一路以来的陪伴与支持。\r\n但由于多方原因，我们不得不遗憾地向大家宣布，《魔兽世界》将于2022年12月24日8:00起正式停止运营。\r\n即日起至停止运营服务前，玩家们仍可正常进行游戏。从今日起，每周四10:00我们将给在线发放代金券作为陪伴礼，玩家可前往邮件领取。我们由衷地希望这段携手共度的时光，能够给各位玩家留下美好的回忆。\r\n具体关服相关事宜流程安排时间表如下，请各位玩家留意：\r\n1. 2022年11月20日8点，停止游戏充值和新用户注册；\r\n2. 2023年1月23日凌晨0点，正式停止游戏运营、关闭游戏服务器；\r\n3. 2023年2月17日下午2点，社区关闭、官方微博账号关闭、游戏官网下架。\r\n*请注意：游戏服务器关闭后，玩家无法再次登录游戏，游戏内的所有帐号数据及角色资料等信息将被全部清空。\r\n对于《魔兽世界》停止运营给玩家造成的不便，我们表示深深的歉意与遗憾。终止线上运营后的其他相关补偿方案，我们将会在后续的正式补偿公告中详细说明，还请各位玩家留意后续公告。如您对停运事宜有任何疑问或建议，可以v我50边请我一顿KFC边向我反馈。\r\n再次感谢玩家们一路以来的支持与厚爱。\r\n",
	"网易和暴雪谈崩，代理商位置空缺，现在是千载难逢的好机会，和我融资一个亿拿下代理，一起发财！现在V我50，我们肯德基详谈融资事宜",
	"网恋靠谱吗？\r\n当虚拟主播认识了一位v，我们聊了大半年，一直聊的都很好，突然她提出要见一面，我犹豫了说我是残疾人，她说她不介意，只做普通朋友，于是我答应了，也要了她的塔照，普普通通的一个人，看起来不讨厌，她离我的城市不远…我们约在海边见面，没有选择商场怕遇到林忆宁，我先到的因为她要开车好几个小时，等待时看着照片以为会是“翻车现场”因为手机都有美颜功能，但是当她走出停车场时我一眼就认出她了，她本人比照片上好看多了，也是我特欣赏的类型，她也认出我了，我不知道该说什么就站在那里，为了打破尴尬，她提议抱一下吧，我说好，但是她很礼貌的摸了摸我头喊了一句离岛姐姐。我一下子就脸红了，心怦怦的直跳，沿着海边走了会我们去吃了饭，她基本没吃多少，整个过程中都在给我剥虾皮，挑鱼刺，吃完饭走出饭店她牵起了我的手，我也没有挣脱开，她说送我回家，怕出来太长时间不播不好，我说我会飞回去很快，她执意要送，把我送到家附近公交站台，下车时她说喜欢我，从来没有当面被表白，当时我有点懵，有点感动，v我50继续给你讲\r\n",
	"⣀⣆⣰⣒⣒⡀⢀⠔⠠⠤⡦⠤⠄⢴⠤⠤⠤⢴⠄\r\n⢰⣒⣒⣒⣲⠄⠠⡎⠸⠽⠽⠽⠄⠼⡭⠭⠭⡽⠄\r\n⢸⠒⠒⢒⣺⠄⠄⡇⡍⣝⣩⢫⠄⣊⣒⣺⣒⣊⡂\r\n⢠⠤⠴⠤⠤⠄⢐⢔⠐⠒⡖⠒⠄\r\n⣹⢸⢍⢉⢽⠄⢀⢼⠠⠤⡧⠤⠄\r\n⡜⡸⠔⠑⠜⡄⠠⡸⢀⣀⣇⣀⠄\r\n⢰⣒⣒⣒⣲⠄⠠⡦⢴⠄⡖⢲⠄⡖⢲⠒⢲⠒⡆\r\n⢸⣒⣲⣒⣚⠄⠄⡯⢽⠄⣏⣹⠄⡇⡸⠄⢸⣀⡇\r\n⣑⣒⣺⣒⣒⡀⢈⠍⠩⣡⠃⣸⠄⣏⣀⣀⣀⣀⡇\r\n⡄   ⡄⠐⢲⠒⠄⡆⠢⠄⡤⠤⠄⢀⠤⢄\r\n⢱⢰⠁⠈⢹⣉⠉⡏⡍⠄⠗⠒⡄⢸    ⢸\r\n  ⠇   ⠈⣹⢀⡠⠺⡰⠄⠢⠤⠃⠘.. .⠜\r\n",
	"你好，我是樱井孝宏，近日有人称我有长达十年的婚外情，我想澄清这完全是误会。这件事严重影响了我的日常生活和职业生涯，所以我必须去把这件事处理好，由于总是来回奔波，今天出门忘记带钱了，现在肚子很饿，v我50，等我吃饱了我继续解决这件事，事情结束后给你们每个人都可以随便把声优和角色联系在一起，我不会有意见的。",
	"你们看微博热搜了吗？李易峰居然放出来了，我一个朋友在北京的派出所上班，说李易峰本来已经开始劳改了，结果今天没去，旷工了，因为峰旷星期四",
	"本人奥托阿波卡利斯在此澄清，关于本人消失于虚数之树的言论皆为无中生有的中谣，不过就实际情况而言，目前本人确实在突破虚数空间回到泛时间线的途径上遇到了一点小小的阻难。不用担心，我的老朋友们，只需要现在向这个账号上打上50块钱，起到一点小小的推波助澜之功，我就有方法立刻重塑通往泛时间线的虚数通道...当然了，我的老朋友们，我奥托向来不是轻义之人，你若予我以滴水，我定能报之以涌泉，只需要v我50，你将来在天命的地位必然是居高而上万人俯首，怎么样我的老朋友们，我认为这笔交易是划算的，你们，是应该相信我的\r\n",
	"有没有人和我一样很讨厌疯狂星期四的文案，尤其是朋友一发就是发一堆，后面还有人跟着发哈哈哈哈，我不知道笑点在哪里，但是我又不好意思叫他们不要发，然后我就想到一个办法，v我50今天肯德基疯狂星期四，我边吃边告诉你这个办法是什么",
	"优秀是“渐渐的事”\r\n成长是“天天的事”\r\n学习是“坚持的事”\r\n别忘了今天是肯德基疯狂星期四",
	"这朋友你到底打不打算处了？我说的是谁自己心里清楚 今天疯狂星期四 香辣翅尖9.9十五根＋鸡翅十块39.9＋葡式蛋挞29.9八个 你还有机会挽救我们的友情",
	"我不是南方姑娘的花裙子，不是傲寒在稻城快要融化的早晨，不是董小姐手中的那支兰州，不是郑州冬天的那缕阳光，不是杂货店老板娘手中的玫瑰，不是北方女王四川路过的江成都见过的湖，不是低苦艾的候鸟飞到的北方，不是祝星踩着的山河，可今天是肯德基疯狂星期四，谁请我吃?",
	"…进化…生存……优先种族需要进食………ۦۥۥۦٌۦٌۥۥۚۚۥۥۥۥۥۥۥۥۥ ۥۥۥۥۥۥۦۦۥۥۦۥۥۥۥۖۛۛۦۦۦۦۦۦۦۦۦۥۥۦٌۥۥۚۚۥۥۥۥۥۥۛۚۚۥۥۥ  ۥV我50…～..**ۥۥۥۥۥۦۦۥۥۦۛۚۗۥjjjjjjjjjĵjjۥۥۥۖۛۛۦۥۥۦٌۥۥۚۚۥۥۥۥۥۥۥ肯德基…ۥۥ ۥۥۥۥۥۥۦ疯狂 ۥۥۦۦۥۥۦۛۚۗۥjjjj星期四ィかぇぃカ■■头七ۥۥۥۥۦۦۥۥۦۥۥۥۥۖۛۛۦۦۦۦۦۦۦۦۦۥۥۦٌۥۥۚۚۥۥۥۥۥۥۛۚۚۥۥۥ  ۥV我50…～..**ۥۥۥۥۥۦۦۥۥۦۛۚۗۥjjjjjjjjjĵjjۥۥۥۖۛۛۦۥۥۦٌۥۥۚۚۥۥۥۥۥۥۥ肯德基…ۥۥ ۥۥۥۥۥۥۦ…进化…生存……优先种族需要进食………ۦۥۥۦٌۦٌۥۥۚۚۥۥۥۥۥۥۥۥۥ ۥۥۥۥۥۥۦۦۥۥۦۥۥۥۥۖۛۛۦۦۦۦۦۦۦۦۦۦۦۦۦۦۦۦۦۥۥۦٌۥۥۚۚۥۥۥۥۥۥۛۚۚۥۥۥ   ۥV我50…～..**ۥۥۥۥۥۦۦۥۥۦۛۚۗۥjjjjjjjjjĵjjۥۥۥۖۛۛۦۥۥۦٌۥۥۚۚۥۥۥۥۥۥۥ肯德基…ۥۥ ۥۥۥۥۥۥۦ疯狂 ۥۥۦۦۥۥۦۛۚۗۥjjjjぃカ■■头七疯狂 ۥۥۦۦۥۥۦۛۚۗۥjjjj星期四…ۥۥ ۥۥۥۥۥۦۦۥۥۦۛۚۗۥjjj\r\n收",
	"新赛季王者代打\r\n\r\n1.大国标马超16000\r\n2.大国标露娜15000\r\n3.大国标澜14000\r\n4.大国标李元芳13000\r\n5.大国标公孙离12000\r\n*6.代吃肯德基50\r\n*7.代吃麦当劳30\r\n*8.代喝奶茶20\r\n目前只接6-7-8 别的不会",
	"你好，金色的那菈，我是兰德基，我听兰阿帕斯、兰萨卡、兰茶荼、兰非拉他们说，你愿意为无忧节做准备，我现在正在进行「烹饪修行」，肯德基，疯狂星期四，是开心的味道，​V我5​0，让新的梦想永远不被无留陀侵蚀，旧的故事与无留陀一同被忘却，森林会记住好那菈的善举",
	"肯德基给你们多少钱了，让你们周周给编段子，疯狂星期四真的那么有诱惑力？我偏不信，v我50，我去试试",
	"我刚刚理性分析了一下，我个人觉得李易峰pc是假的 因为走完整套流程就是周四了 有没有可能其实李易峰是帮吴亦凡带肯德基进去​",
	"正确约炮的标准流程：一般都上午10点以后，女的打扮打扮，男的取点钱，见面了，女的假装不好意思，男的一般都夸女的比照片好看，然后问女的爱吃啥，就领吃饭去了，男的开始展示自己的幽默然后溜达，看电影，一晃一下午女的也不说回家。男的寻思这是没跟我呆够啊，就说咱俩唱歌去吧，女的说太晚了，男的说没事一会我送你回家。一般这阶段男的不会把女的咋地，这时候时间也有点晚了，男的说走吧我送你回家，女的还得寻思，这男的是没看上我啊，心里有点小失落两人心里就开始寻思怎么能不回家。男的说这么晚了，你回家我也不放心啊，我有点没待够，女的说我必须得回家，男的说我送你，结果两人墨迹墨迹的10分钟没走几米。后来女的说我回不去家了，理由可能是寝室关门了，或者朋友对象回来了，男的一本正经的说，要不我给你找个宾馆吧，女的说，不过你得答应不碰我，男的说，放心吧，我不是坏人，然后微微一笑深藏功与名。然后两人就进宾馆了，先聊会天看会电视，男人时不时的说女的，你长的真好看，男的说，你明天还上班或者上学，早点睡吧，女的还得装不好意思，男的说我能抱着你睡么，什么也不干。女的说行，男的就狠狠的抱着女的说，亲爱的，我这么抱着你，你就不冷了，女的也抱上了，男的说我可以亲你吗，女的也默认了，可是一般这时候还不会发生啥。这时候男的想有进一步的动作，开始脱女的衣服，脱到底线，女的这时候开始拒绝，推开这个男的，男的一看情行不行，就说宝贝对不起，我太冲动了，你要知道我是个男的，何况你还长的这么好看。女的这时候就把男的当亲人，讲述自己情路坎坷，人生迷茫，男的这时候就开始展现自己的温柔，宝贝你要是早遇见我就好了，一定不会让你吃这么多苦，你别怕，还有我呢。突破了女的的心灵防线，大家就都放的开了，女的开始亲男的，然后男的翻身小声说道：今天肯德基疯狂星期四可以请我吃嘛？",
	"肯德基这逼养的，怎么又到了周四，这一次，肯德基疯狂星期四，我一定要吃，我本是显赫世家的少爷，却被诡计多端的奸人所害！家人弃我！师门逐我！甚至断我灵脉！重生一世，今天肯德基疯狂星期四！谁请我吃？家人们，别他妈垂头丧气了！知道今天是什么日子吗？今天是肯德基fucking crazy Thursday！吮指原味鸡10块钱2个，家人们v我100，我他妈要吃20个！",
	"姐妹抱抱你，我看了也很气，这件事闹得挺大的，但也不是特别大，你要说小吧，倒也不是特别小，我觉得这事还是挺大的，不过不是特别大，但也不小。大家都觉得这事特别大，但我觉得也没那么大，但你要说小吧，也不小。到底是啥事儿，v我50，我今天亲自去肯德基看看。",
	"某个人，不回消息永远别回了，到底群消息重要还是我重要，整个群我只对你一个人有感觉，难道你心里就不明白吗？不然我整天闲得来这里聊天，我不会跑别的地方聊天玩吗？你以为我天天闲得慌吗？我如此的喜欢你，今天肯德基疯狂星期四葡式蛋挞29.9八个，你还有机会挽回",
	"全员核酸检测通知 明日（9月29日，周四）本群进行全员核酸检测，检测时间、地点安排如下：\r\n 一、人员：所有人 \r\n1.检测时间：17:00—18:30 \r\n2.检测地点：肯德基大门口 \r\n二、 要求 \r\n1.戴口罩，保持一米线，自觉排队，准备好微信扫一扫，v我50后迅速离开。  \r\n2.请认真核对人数，确保不漏一人",
	"临江仙•抒怀\r\n疯语薄言悲复喜，\r\n狂了营营此生。\r\n星低云阔夜流空。\r\n期年终碌碌，\r\n四季苦匆匆。\r\n请君莫笑当年事，\r\n微名何羡留声。\r\n我自孤行尘寰中。\r\n五陵芳华尽，\r\n十里散秋风。\r\n",
	"╭◜◝ ͡ ◜ ╮ \r\n(    好想    ) \r\n╰◟  ͜ ╭◜◝ ͡ ◜ ͡ ◝  ╮\r\n　 　 (  有人v50   )\r\n╭◜◝ ͡ ◜◝ ͡  ◜ ╮◞ ╯\r\n(   请我吃KFC  ) \r\n╰◟  ͜ ◞ ͜ ◟ ͜ ◞◞╯\r\n╭◜◝ ͡ ◜ ͡ ◝  ╮\r\n(  有人v50   )\r\n╭◜◝ ͡ ◜◝ ͡  ◜ ╮\r\n(   请我吃KFC  ) \r\n╰◟  ͜ ◞ ͜ ◟ ͜ ◞◞╯\r\n",
	"从前有一个国王叫肯，娶了一个歌姬为妾。国王的国家矿产资源发达，国王十分宠爱歌姬，将一部分矿产给了歌姬的家族开发。但歌姬十分贪婪，为了实现矿产垄断，歌姬把其他同行的矿井都给封了，包括国王分派给贴身武士的。于是国王把歌姬抓起来审判，歌姬问定什么罪？国王说：死罪。肯的姬封矿刑期死，为我武士。",
	"Python大佬，我写了一段代码突然报错了，找不到原因\r\nTraceback (most recent call last):\r\nFile \"C:\\\\python\\kfc.py\", line 8, in <module>\r\nKfcError: KFC Crazy Thursday WhoEver Transfer Me 50 CNY Via WeChat, I Will Thank Him.",
	"接代打\r\n188块一天内上王者 如果没上王者退138块 留50辛苦费 有需要的直接给50块就可以了 号也不用给我 反正打不上去 省的退钱了 麻烦要打快点 我马上要路过KFC 店了",
	"个人副业，支持一下\r\n1. Python 脚本 20元/100行 \r\n2. Java 程序 500元/个，复杂度面议 \r\n3. 网站及 APP 开发 2000/个，复杂度面议\r\n4. AE/PR 视频剪辑，简单的 80/个，复杂面议\r\n5. 音频处理，移调加无损音源 15/个，消除人声等处理面议\r\n6. 鼓组编排，80/首，150两首。风格面议\r\n7. 高达模型代拼（水口精修 600 - 2500 依次 + 渗线，水贴，不包模）sd50，hg130，rg200，mg300，pg面议\r\n8. 摄影＋后期，30/组\r\n9. KFC 代吃 50/次，只接周四 \r\n\r\n目前只接 9 ，其他的都不想接",
	"群里的某个兄弟 ，我就不指名道姓了，做人别太过分了，真的，我不想撕破脸，你在背后这么绘声绘色地说我，如果不是别人跑来亲口跟我说，我还真不知道你这么多嘴，大家都是因为朋友认识的，做事真的不要太过分，换位思考下如果我也这么说你你会怎么想，话今天我撂这儿了，我把你当兄弟，有什么事情和你分享。你居然背后告诉别人我一晚能干8次，还说我像打桩机，姿势多，力道猛，搞得后来好多女的加我，最近身体已经处于崩溃边缘，实在受不了了，今天是疯狂星期四，快v49请我吃肯德基让我补充一下身体！",
	"一周有7天：MON，TUE，WED，KFC，FRI，SAT，SUN",
	"亲爱的展博：当你看到这封信的时候，我已经在去机场的路上了，请原谅我改签了航班，我知道如果我看到你们肯定就走不了了，因为我舍不得你们每个人，尤其是你，和大家在一起的日子，是我一生中最快乐的时光，虽然我也不想结束，但是新的故事总要开始，展博过去我不懂爱是什么，是你让我明白，爱是当你爱上一个人会舍弃自己的自由换取他的自由，爱是当你爱上一个人会改变自己的人生，成全他的心愿，爱是当你爱上一个人会愿意放开手，留下最好的回忆和祝福，爱情最美的，不一定是终点，旅途一起走过，也以不负一生，原谅我的天真，这是我能想到的，最好的结局。如果你想我了，请v我50一起吃顿炸鸡。",
	"我其实真的很想你会回来 我在深夜想起你的时候还是会忍不住掉眼泪 消息写了又删 我也在你离开的时候拼命挽回过 这辈子最大的遗憾就是没能和你走到最后 答应过你的好多事都还没有完成 要是能在结婚的年纪遇见你就好了 我觉得这个世界好不公平真心总是被辜负 我求你回来的样子一定很烦吧 这些天写了很多话 想着哪天给你看看 想让你知道我这些天怎么过来的 你以前问我跟你在一起后不后悔 说实话我没有后悔 因为只要是我喜欢的不管好不好我都会把我最好的给你 我已经习惯有你的每一天 你走了 离开了 不要我了 我每天就像个废人一样抽烟喝酒 我以为这样就能消磨对你的思念 可是我每次喝醉了想到的人是你 可能觉得我特别没出息吧 我们不是不合适 每个人的性格不一样互相磨合互相体谅互相包容 有些时候吵架和你冷战 不是因为我不爱了而是当时的我不知道怎么去跟你沟通 我没有那个能力 我不知道怎么去说去解释才能让你明白我的意思 我不知道怎么做才能和你的想法达成一致 往往好多时候 我们吵架就愈演愈烈了 每次吵完都会后悔好久 为什么当时就不能理智一点 控制自己的脾气不和你争吵 为什么不多包容你 我其实心里会想很多 我不说我比别人爱你 我只能说只要你一句话不管多远我来会来见你 你让我做什么我都满足你 你走后我的世界就剩我一个人了 心里空空荡荡的 整天做什么失魂落魄的 失去的不会再回来 错过的终究不会再遇见 大概不打扰才是最后的温柔 毕竟有些感情除了说再见别无选择 本以为陪我到最后的人一定是你 所以我把自己全部的感情全部投入到你的身上 一切时间一切精力都放在你身上 爱的太满结果陪我一生的人却不是你 你知道那种感觉吗 每天晚上我都会胡思乱想很多以前有你的时候 想起我们以前快乐的事 我真的很脆弱你走了我的世界都要崩塌了 夜里睡着突然被惊醒的那种感觉那种快要窒息的感觉 我比谁都喜欢你 但是没有用啊 我不知道这些天是怎么熬过的每天丧到极致做什么都会想到你 每天沉沦在明明知道没希望还要无尽的等待整夜失眠 我每天都幻想着你会来找我 很想知道你的近况 爱而不得的感受真的太难受了 很庆幸我能走到你的身边 却没能给你一个家 以前老是说会保护你会给你安全感 后来你说我不是你想要的 原来真的有两个人 要相互喜欢 感谢你出现再我的生命中 跟你在一起的那些时间我真的特别的开心  所以今天可以微信转给我五十块钱吃肯德基吗",
	"我是羊了个羊的游戏设计师 今天我被公司开除了 因为我掌握着第二关的通关密码 所有人都追着我 我现在无处可藏 只能向你求助 今天疯狂星期四 你V我 50 我就把羊了个羊第二关的通关秘籍传给你",
	" 如果你现在买iPhone14和一块苹果手表。你就不会有朋友。其他人都会只会尊敬你，和你说客套话。你永远不可能和他们嬉笑打闹。除了礼貌的微笑，你得不到任何东西。你会孤独的走进坟墓。没人陪你。这就是我现在不买iPhone14和苹果手表的理由。 但如果有朋友v我五十块让我去肯德基疯狂星期四，我们就会建立起最纯真最坚固的友谊，收获一片真心。所以你现在买iPhone14和苹果手表不如v我50一起疯狂星期四。​\r\n",
	"-iPhone 14：5999起\r\n-iPhone 14 Plus：6999起\r\n-iPhone 14 Pro：7999起\r\n-iPhone 14 Pro Max：8999起\r\n-Apple Watch 8：2999起\r\n-Apple Watch SE 2：1999起\r\n-Apple Watch Ultra：6299\r\n-AirPods Pro 2：1899\r\n-V我 ：50起 ​​​\r\n",
	"14不香了，果然还是13更香，iPhone14真是更新了个寂寞！你看过iPhone14发布会了吗？多少会有些失望，iPhone14真是更新了个寂寞啊，除了eSIM卡和卫星通讯两张无关痛痒的更新，其实纯粹就是数字变化，从13到14，采用了13相同的A15芯片，相同尺寸的屏幕，这也能叫全新一代的iPhone14手机，这无疑是妥妥的智商税！近期准备换机的那些人，反正我不建议购买iPhone14，不如把钱拿来放在填饱肚子上，今天肯德基疯狂星期四，谁请我吃？",
	"没关系我不缺钱\r\n是我对不住你\r\n没有没有燕子幸亏咱们没领证，领了证，我耽误你一辈子\r\n我走啦\r\n\r\n你保重啊\r\n你也保重\r\n再见🚖🏻‍♀🏻‍♂\r\n再见还会再见吗🥺🥺🥺燕子，再见的时候你要幸福好不好，燕子你要开心😧😧你要幸福好不好，开心啊🥺幸福🤧🤧\r\n你的世界以后没有我了没关系你要自己幸福🥺🥺🥺\r\n燕子\r\n燕子😫😫😫🚖～～\r\n燕子没有你我怎么活啊……燕子…燕子…燕子 🧎🏻没有你我还怎么活啊\r\n群友们 v我50，我去打车追",
	"我不知道什么文案，不知道什么kfc，也许今天是疯狂星期四但你不会v我50，但你要记住:今天是ame水人波高地四周年纪念日！",
	"肯德基这 beyond 的，怎么又到了周四，这一次，肯德基疯狂星期四，我一定要吃，我本是显赫世家的少爷，却被诡计多端的奸人所害！家人弃我！师门逐我！甚至断我灵脉!重生一世，今天肯德基疯狂星期四!谁请我吃？家人们，别他妈垂头丧气了 知道今天是什么日子吗？今天是肯德基fucking crazy Thursday！吮指原味鸡10块钱2个 ，家人们v我100，我他妈要吃20个！",
	"今天，是我人生中最黑暗、最emo的时候，但是我很难把朋友当成垃圾桶或者树洞去倾诉，更喜欢憋着自己消化，睡一觉第二天醒来就好了，不想连累别人，不想把负面情绪传递给身边的朋友。但是，我更希望家人能够给予我物质层面的帮助！今天是肯德基疯狂星期四，有哪个家人愿意请我吃肯德基吗？",
	"姐我再说一遍，你是我姐，你不是粉丝，你不要再让我听到，你说你是粉丝，这两个字我不要再听到了，听到没有，我再给你说一遍，如果你再敢质疑这个，就v我50吃肯德基疯狂星期四",
	"\\u6211\\u672c\\u662f\\u663e\\u8d6b\\u4e16\\u5bb6\\u7684\\u5965\\u7279\\u66fc\\uff0c\\u5374\\u88ab\\u8be1\\u8ba1\\u591a\\u7aef\\u7684\\u602a\\u517d\\u6240\\u5bb3\\uff01\\u5965\\u7279\\u66fc\\u5bb6\\u65cf\\u5f03\\u6211\\uff01\\u5965\\u7279\\u4e4b\\u7236\\u9010\\u6211\\uff01\\u751a\\u81f3\\u65ad\\u6211\\u4f3d\\u9a6c\\u5c04\\u7ebf\\uff01\\u91cd\\u751f\\u4e00\\u4e16\\uff0c\\u4eca\\u5929\\u80af\\u5fb7\\u57fa\\u75af\\u72c2\\u661f\\u671f\\u56db\\uff01\\u8c01\\u8bf7\\u6211\\u5403\\uff1f",
	"你想象一下，过几天就要七夕了，你一个人单着身刷着抖音，你的兄弟姐妹们都换上了情头，给你讲甜蜜爱情历程，就连打游戏都会发现一堆情侣秀恩爱，你是否会后悔今天没有找我下单!\r\n\r\n换情头：20元\r\n\r\n秀恩爱：30元\r\n\r\n陪聊天：40元\r\n\r\n猛汉王陪玩：50元\r\n\r\nszb坐牢：88元\r\n\r\n全套：200元\r\n\r\n不限年龄 我不是复制我是真业务七夕节跟谁过最幸福?\r\n\r\n口舔狗 口渣女 口海王\r\n\r\n口炮王 口渣男 √我\r\n\r\n注：让我滚的话转我58我要吃两份吮指原味鸡，退订请v我50回复TD",
	"有些事你不了解，不能乱说。确有此事，我通过我家族世代流传的石碑上看到的，此阵法是用来打开连接现世和混沌之间的大门，现在他们要在2035年发生日全食的时候解除封印让魔王德古拉复活毁灭现实世界。\r\n\r\n我说的都是真的，因为我姓贝尔蒙特，v我50助我重铸吸血鬼杀手，通关致谢名单就有你的名字。 ​​​",
	"KFC订餐系统存在业务逻辑漏洞(CNVD-2022-0728)\r\n★关注(999)\r\nCNVD-ID CNVD-2022-0728\r\n公开日期 2022-07-28\r\n危害级别 高(AV:LAC:LAu:N/C:C/E:CIA:C)\r\n影响产品 影响所有KFC订单系统\r\nCVEID CVE-2021-0728\r\n漏洞描述 肯德基(Kentucky Fried Chicken，肯塔基州炸鸡，简称KFC)，是美国跨国连锁餐厅之一，有着庞大的用户群体。\r\nKFC的点餐系统存在逻辑漏洞，可让用户使用大额优惠购买其产品。造成巨额的经济损失。\r\n漏洞类型 通用型漏洞\r\n漏洞报送者 txf\r\n漏洞解决方案 目前厂商无法修复漏洞，且漏洞有周期性和时效性。具大量数据分析，该漏洞可能在每周四就会被自动触发。\r\n厂商补丁 无补丁\r\n验证信息 已验证\r\n报送时间 2022-07-27\r\n收录时间 2022-07-28\r\n更新时间 2022-07-28\r\n漏洞附件 附件暂不公开\r\n在发布漏洞公告信息之前，CNVD都力争保证每条公告的准确性和可靠性。然而，采纳和实施公告中的建议则完全由用户自己决定，其可能引起的问题和结果也完全由用户承担。是否采纳我们的建议取决于您个人或您企业的决策，您应考虑其内容是否符合您个人或您企业的安全策略和流程。\r\n(编辑：CNVD)|已有0条评论\r\n",
	"Steam上多买了一个艾尔登法环的key，送给有缘人了:KFCC-RAZY-THUR-SDAY-VME50",
	"感觉大家好有趣啊，不像我，不仅连句话都搭不上，还要被当成破坏氛围的傻狗,我现实生活中自闭没朋友，哪怕是在网上也受尽冷眼，每次组织了半天的语言都如鲠在喉，最后还是默默删掉了看你们互动，融入不了群体的我还是想偷偷问：今天是肯德基疯狂星期四，能有好心人请我吃吗？",
	"假如你是李华，你的英国笔友Peter给你写了一封信，向你询问你周四的安排，请你写一封回信。\r\n写作要点:\r\n1.你要去肯德基参加\"疯狂星期四\"活动。\r\n2.向他介绍\"疯狂星期四\"的活动。\r\n3.要求他V你50。\r\n注意事项:\r\n1.行文流畅，包含所有要点，无语法错误，可适当自由发挥,80词左右\r\n2.不出现真实的人名地名。\r\n3.请将Peter的50 V给我，否则不给分。\r\n\r\nDear Peter\r\nI will introduce my plan in this Thursday for you，I will go to KFC and have__________________________________________________________________________________________________________________________________________________________________________________________\r\nI'm looking forward to your reply and the V 50 from you.\r\nYours\r\nLiHua",
	"穿过挪威的森林，让我走进你的梦里，夕阳落在我的铠甲，王子不一定骑白马，现在是下午三点四十六分，你不必故作冷淡，我也不想做最后的纠缠。最后问一句，疯狂星期四，请我吃个肯德基？",
	"你觉得这个群有什么问题？\r\n ○违法违禁\r\n ○低俗\r\n ○赌博诈骗\r\n ○血腥暴力\r\n ○人身攻击\r\n ○青少年不良信息\r\n ●疯狂星期四群主竟然不请群员吃肯德基\r\n ○有其他问题",
	"作为一个恋爱老手，给大家的九条建议:\r\n1.谈恋爱首先要找你爱的，如果结婚就要找爱你的\r\n2.千万别输在“等”这个字身上\r\n3.永远留住30%的神秘\r\n4.不要低估任何一个人\r\n5.别把没教养当做有气场\r\n6.谈恋爱可以穷，结婚不可以\r\n7.谈恋爱一定要找我\r\n8.v50请我吃肯德基疯狂星期四\r\n9.牢记第8条，前7条没什么用\r\n",
	"今天要一起去求姻缘吗？有一家新寺庙，名字叫肯德基疯狂星期寺。",
	"公司服务器被黑客攻击，我成背锅侠了。因为昨天上线了新功能，他们一致认为是我的锅，明眼人都知道，我只是一个小白，根本不可能导致服务器被黑。哎，想到这里我不自觉流下了泪水，4月-6月的工资还没发给我。现在新锅又来了，他们是打算彻底不发工资了我？刚才人事还说，因为服务器被黑，要我补公司5万元。我吃饭钱都没有了，还要倒贴给公司？我想了程序员经典名言：程序和人哪个可以跑？跑确实让我遗忘了一切，在跑的路上我起了每周四肯德基都有疯狂星期四，我决定今晚就去狠吃一顿，抚慰我最近倒霉的遭遇。希望有缘人V我50，办法总比困难多，出路都是自己走出来的，行路难，行路难，多歧路，今安在？长风破浪会有时，直挂云帆济沧海。。。。。。",
	"⢠⠤⠴⠤⠤⠄\r\n⣹⢸⢍⢉⢽⠄\r\n⡜⡸⠔⠑⠜⡄\r\n\r\n⡢⡂⠒⢲⠒⠂\r\n⡠⡇⠤⢼⠤⠄\r\n⢄⠇⣀⣸⣀⡀\r\n\r\n⢰⣒⣒⣒⣲⠄\r\n⢸⣒⣲⣒⣚⠄\r\n⣑⣒⣺⣒⣒⡀\r\n\r\n⢴⠤⡦⢰⠒⡆\r\n⢸⠭⡇⢸⣉⡇\r\n⡩⠉⢍⡜⢀⡇\r\n\r\n⡖⢲⠒⢲⠒⡆\r\n⡇⡸⠄⢸⣀⡇\r\n⣏⣀⣀⣀⣀⡇\r\n\r\n",
	"你跟你男朋友奔现进房间以后，裤子一脱你花容失色的质问你男朋友：你不是说你有18cm吗？怎么这么小？你男朋友说：因为今天是肯德基疯狂星期四活动 满18减15",
	"疯狂星期四都买不起？有手有脚不会自己工作劳动挣钱？用自己劳动换的血汗钱用着不都心安一点？非得那么卑微去问别人要？觉得我说的对的给我60  说的有点累了 买点肯德基吃一下",
	"被群内渣女欺骗四年，说进群就分配富萝莉，但是时至今日，群里面还都是一群和我一样没人要的骚话网友，我很心痛天天以泪洗面，最近没有怎么哭了，慢慢变好了……以前有多快乐，现在就有多难过。从听到分配富萝莉的快乐，到被欺骗的委屈，用真心换群主的欺瞒，很痛，也很难。今天是肯德基疯狂星期四，v我60，抚慰我支离破碎的心，别问我为啥比他们多10，我贪心想多喝杯杨枝甘露",
	"7月求姻缘应该去哪个寺？\r\nA、浙江省杭州灵隐寺\r\nB、广东省深圳弘法寺\r\nC、江苏省江市甘露寺\r\nD、肯德基疯狂星期寺",
	" 昨天公司新来一位女同事今天她找我聊天\r\n\r\n她说她是她爸妈捡来的\r\n\r\n现在19岁，不是亲生的，她自己也知道。\r\n\r\n她哥今年27了，读研读博，所以现在还没谈女朋友，她妈突然就跟她说，等她毕业了，她哥要还没有对象她就跟她哥结婚吧。她当时还在看书听到这话吓得我魂都丢了，果断说不行，她妈就说先别急，听她讲完。\r\n\r\n她要和她哥结婚了，不用担心她哥对她不好，第ニ她哥哥也不会有婆媳矛盾，第三也不会因为任何原因离开她离开家。\r\n\r\n第四她不用养双方父母，将来爸妈生病了她们可以一起照顾，第五知根知底她哥哥还没谈过对象是干净的。\r\n\r\n她讲完她就沉默了，确实除了不相爱以外全是利没有弊，她找不出任何驳的理由，本来想说她们没有那种感情只有亲情，她妈后来就说她以后也不一定就能遇到爱她爱的死去活来的人，大家相亲结婚不就是奔着凑合着过的念头才在一起的吗？\r\n\r\n她说决定复仇，但是肚子太饿，刚好今天是肯德基疯狂星期四， v 我50我给她肯德基吃，到时候我们一起听她的复仇计划",
	"明天周四了 又到了令人无比激动又十分难得的KFC疯狂星期四 对于明天的疯狂单品我很是期待 但是我是穷逼 我实在是想请我家亲爱的穷逼祝祝奢侈一把 谁能满足一个穷逼为了另一个穷逼的小小愿望呢~",
	"记得16岁那年，第一次和同桌接吻，快亲上的时候，她突然说等一下，我就纳闷了她要干嘛？只见她小心翼翼地从兜里拿出三个糖，有草莓苹果和荔枝味的，她让我挑一个最喜欢的。我指了一下那个荔枝的，然后问她干嘛?她二话不说撕开糖纸，就把那颗糖给吃了，然后一把扯过我的脖子，我俩就接吻了，全程一股荔枝味，后来她跟我说，人生那么长，我没有自信能让你记住我，但是你既然喜欢吃荔枝味的糖,我只能让你记住和我接吻的时候是荔枝味的，这样以后你吃荔枝味的东西都能想起我，我和你接吻的味道。如今我们分手好多年了，每次吃荔枝味的东西都会想起她，家里固定有荔枝糖，想她了都会吃上一个，就好像在和她接吻。若还有机会真想告诉她，人生那么长我可能要记着你一辈子了。后来，我有过两个女朋友，也终没有结果，时间就这样沉淀下去，终于有一天，我再也无法抑制我心中的那份情感，我决定去找她，我们要在一起，后来经多方打听才知道，她毕业后找了份不错的工作，工作几年后，毅然辞职自己开了家糖果店，而我终于有一天找到她，开口的第一句：还记得那次荔枝糖的味道吗?她强忍着泪告诉我，荔枝糖的味道她一直没忘记，只是我们再也回不去了。我没有转身离开，也没有奋不顾身的冲上去抱住她说出多年来心里一直只想对她说的那些话。就这样，我们傻傻地看着对方，彼此沉默了很久。夕阳的余晖透过窗户斜映在她的脸庞，一如当年那般美里，突然心里流过一股暖意，仿佛那些年曾一起走过的旧时光还在脑海里挥之不去。或许，这已经足够了。有些人，有些事， 一旦错过了就是错过，不再擦肩，也不再回头。虽然岁月带走了我心中最美好的曾经，但岁月带不走的是我那颗永远爱你不变的心 。打开手机准备翻找我们的曾经。不小心打开了肯德基，想起来今天就是疯狂星期四了，所以说谁请我肯德基？吃完我继续说。",
	"记得去年我在掘金认识一个女生，她开始问了一个Vue的问题，说了半天也没说明白问题，群里没有人理她，然后我让她贴代码，代码贴出来也贴错了，根本没贴关键代码，然后几经周折解决了。 然后又有一次她遇到问题，又贴了出来，我刚好写完bug，看了一眼群里，给她解决了。然后她加我了，我说怎么了，她说群里的人都不理她，我说也奇怪了，为啥不理你呢，她说她也不知道，她说要不以后我问你吧，我说可以呀，我看了一眼她的朋友圈，我草，真好看，笑起来，像一个躺着的括弧 ：）。 我们就这样有bug没bug都会聊天，我们不断攀谈，唉，何其有幸啊，我遇到了她，我时长感叹，我这样普普通通的人，遇到了这样好看且动人、可爱的女生。 那段时间，兄弟萌可能在群里很少看到我，是的，我恋爱了，我喜欢上了这个女生，她也被我的真诚所打动，可能就是缘分吧，我们在一起了。 到今天，我们在一起已经两个月了，前两天，我们吵架了，因为我已经无数次和她说变量命名要规范，类名要语义，我就说了她一下，她哭了，我知道从她哭的那一刻起，我已经错的一塌糊涂了。我讨厌我的完美主义和固执，她已经两天没回我信息了，我决定去找她，去她的城市——北京，我刚下飞机，今晚就要见到她了，我想了两个晚上，我已经想好了怎么道歉，怎么哄好她，只要我拿出肯德基疯狂星期四藤椒无骨大鸡柳，香味就可以充满她整个房间！她肯定感动得原谅我，谁赞助我一份？感谢了。",
	"通知：今天周四，不允许发跟肯德基和疯狂星期四的段子，违令者给我也买一份吃。",
	"0约1去开房，0到之后发现床上坐着个陌生人怀疑是不是走错了房间，正要关门时，1刚洗完澡从卫生间出来，0看到后顿时心生不满的质问道这是谁，1连忙走到他的耳边轻声道“今天肯德基疯狂星期四，买1送1”",
	"到底肯德基给了多少钱，每个星期四都在问KFC疯狂星期四有无，还每个星期都编新的段子，我就不信肯德基有这么大的魔力，V我五十，我今天亲自去肯德基看看",
	"找对象五年了，多了87个外卖地址，31个收货地址，别的什么也没留下。时常被人完弄感情，但还是憧憬爱情，想起这些事趴桌子上大哭，正当我打开窗户想跳楼的时候，不小心打开了群，觉得也许还能再相信一次爱情，今天是疯狂星期四 群里帅哥，美女v50，我给你们讲讲我的故事",
	"今天、我人生の中で最黑暗、最黑暗の时刻ですが、わたし倒垃圾や树洞のような朋友も不在、自己自身を消化すること挺好、天亮まで睡って目を醒ます、疲れた别の人にも不想疲れ、周围の朋友に否定的な感情を影响たくない。但かし、我は我の朋友が我に物质的な给我ほしい、今天は肯德基疯狂星期四です、来几个たまで朋友は肯德基を吃べるため请我したいです。仲間に冷たく暴力を振るわれて半年、最近あまり泣かなくなって、だんだんよくなってきた……以前はどんなに楽しかったか、今はどんなに悲しいか。人間花火の日常から、目を赤くして別れを告げ、お互いの世界に消えていくまで、痛くて、難しい。今日はKFC狂乱木曜日,微信V我50,慰撫我支離滅裂的心",
	"今天，这是我一生中最黑暗的时刻，但我发现很难像对待垃圾桶或木洞一样对待朋友，更喜欢消化自己，第二天睡觉醒来，甚至不想让别人感到厌倦，不想把负面情绪传递给周围的朋友。不过，我要朋友给我物质上的帮助，今天是肯德基疯狂的星期四，有朋友想邀请我吃肯德基",
	"分手了，其实从朋友到恋人我们发生了挺多事情，能走到一起也是很不容易。我喜欢你，很喜欢你，也想和你幼稚。没想到我能为你做的最后一件事竟然是走出你的人生，我一点都不后悔，更谢谢你教会我成长!我也不知道这是谁写的，挺感动的我就复制下来了，重要的是有人能请我吃肯德基吗",
	"夫人，最近公司财务很紧张，董事长决定辞退一批员工。您爱人是没有特殊技能的普通员工，您也不想让他失去工作吧？还有健太又在学校里霸凌别的同学了，而且他现在成绩很低，几乎可以说不能毕业了。那么将来尊夫的工作和令郎的毕业问题会怎样发展，就看您今晚和以后的表现了，拜托了！另外夫人您也很想要吧！虽然您的嘴里在说不要，可是身体却很诚实地开始抖动了呢！这么美丽的一个女人，很难想象您竟然会做出这种事情，如果您不想让令尊和令堂看到你现在的丑态，就给我小心一点，以后每个周四把肯德基送到我办公室里！还有这件事是我们两个人之间的秘密，您把最好不要给我告诉任何人！我知道您公司在哪里，也知道您在哪个部门，您最好给我好好记住，懂",
	"离婚这么多年了，竟然在群里遇到你了，你过得挺好我也就放心了。孩子上四年级了，成绩一直都很好，连续三年拿班上第一名，就是性格像你，每次生气就逃避，孩子总问妈妈什么时候来参加家长会！有时间就来看看孩子吧，多陪陪孩子，孩子今天想吃肯德基，因为今天是肯德基疯狂星期四，v我40，我请孩子吃",
	"我是盗号的 我把这个人的号盗了 但是我看了这个人聊天记录发现他过得非常艰苦 他生活过的一直很烂 我希望有人看见了能帮助他 让他能有钱去吃肯德基疯狂星期四 就这样吧 眼眶都湿润了 我下了",
	"早上一过来努力写的代码，运行起来一直报错，找不到什么原因，不知道怎么解决，球球大佬帮我看下！以下是报错信息： java.io.IOException: KFC Crazy Thursday need 50$.",
	"我有时候会觉得大家并不喜欢那个真正的我。在网络上，我总是善于伪装，看起来每天都很快乐，无忧无虑，没有烦恼。我的生活也看起来很简单，没有很多人向往的灯红酒绿，纸醉金迷。我很宅，喜欢打游戏，现实中的我并不像网上这么有趣。我其实话很少，最爱干的事是一个人发呆。这样枯燥的我，真的会被大家喜欢吗？我很疑惑。 如果你们能一层一层剥开我的内心，你们会发现，那颗心里写满了一句话：今天肯德基疯狂星期四，谁请我吃?",
	"已知三角形△ABC AB＝25 AC＝18 sin∠BAC＝0.2 过AB中点K 作KF⊥AB交AC于F ，连接中点K及三角形顶点C，得三角形KFC，求证KFC=V50",
	"我有个朋友出去上了几年学回来后就很奇怪。每个月都会进行几次神秘仪式。会和一帮信徒去一个红色的庙里吃以特定的方式进食禽类的碎尸块，还会放上一些红色和白色的粘稠物质。引用一种不可名状冒着气泡仿佛沸腾的黑色液体。我摸过一次不是热的反而冰手。边吃边念叨着什么。疯狂啊。上校啊。还对4这个数字崇拜什么神赐之日之类的。今天就是他们进行仪式的日子。v我50我去跟踪他现场直播给你们看。",
	"こんにちは、b駅董事長陳叡です。 私は実は李副ceoに架空にされて、今歩道橋の下で漬物をかじっています。 しかし私は彼女に反心があると予想していました、その前に私の無数の忠誠心の部下と二次元たちを会社の各レベルに埋めて、今ただv私50だけで、彼らに1回の狂気の木曜日を食べさせて、彼らを再起動して会社の大権を奪還することができて、b駅を再び二次元の懐に戻すことができて、その時、直接あなたをb駅グラモーガン支部の総裁に命じて、更にあなたに1万年の大会員を送ります",
	"我被阴阳寮开除了 不想肝 御魂很差劲 寮友们都不喜欢我 我的亲友也不管我 现在我在黑夜山下 外面很冷 我一张蓝票也没有 我今天30也没做鬼王也没打 我整个人都晕乎乎的 连口宴会都没得吃 体力也快没了 还不小心点开了肯德基 今天是疯狂星期四 V我88 请我吃肯德基",
	"我是高中生侦探工藤新一。我刚在游乐场被打晕，被黑衣组织强迫灌下了STARS-607，现在身体竟然变成了小孩子，目前我吃了灰原哀开发的解药试作品都起不到作用，现在听说肯德基疯狂星期四的蛋挞和甜筒有特殊作用，希望你能够帮我一忙",
	"大家好 我叫田所 今年24岁 今天下午 我邀请了我一直暗恋 但却迟迟无法坦白的后辈远野来我家 想趁此机会向他坦白 但是我发现没有适当的用餐环境给我俩 正好我通过软手机得知今天是疯狂星期四 肯德基会有折扣 所以...有没有好心人能借我114.5元去肯德基和远野好好表白? 作为报酬 我会给那位好心人一张值14元的淳平雪餐厅折扣眷 (迫切",
	"从事了快4年的安卓开发，被身边的公众号文章轰炸的焦虑的掉头发。看了别人写的技术分析，大佬写的文章，不经感叹自己好像什么都不会。再过俩年就要30岁了，时间真的是不饶人，担心被淘汰，未来的职业规划好像变得不是那么合理。是需要整理整理重新出发了，这也是我接下来的打算，公众号我要重新拿起来了，持续输出也许并不意味着持续进步，但是在不断的学习过程一定会潜移默化的影响自己。无论是输出生活还是学习，每一个阶段的感悟和总结的知识，总会一点点一点点的影响未来的我。我渴望变强。V我50,吃肯德基疯狂星期四,帮助我变强.",
	"我是一名间谍，代号「黄昏」我和我的女儿阿尼亚被尤里赶出家门，现在在外饥寒交迫。阿尼亚要吃汉堡，如果不去她就不会去伊甸学院上学，「枭」计划也会失败。为了维护世界和平请v我50，我要带阿尼亚吃疯狂星期四",
	"刚刚估完分，意料之外却又情理之中的考砸了。言语无法形容我的失落。和一模二模的分数有不少的差距，理想的大学应该是再也没有希望了，与学姐约定好透她的誓言再也无法兑现。满满的期待和信心在对完答案之后的那一刻被磨灭殆尽，耳朵也突然嗡的一下开始了耳鸣。真的很对不起父母的培养以及学姐的期待，也不知道自己是不是应该再来一次。心情真的很糟糕。如果你同情我的话，v我50让我买今天的肯德基疯狂星期四帮我舒缓一下心情谢谢",
	"我有朋友去 Adobe 上班了，他跟我告诉，其实不用花钱，输入内部的序列号就可以免费用正版 Photoshop 2022，我把 key 分享有缘人：KFC-CRAZY-THURSDAY-VME50（懂得都懂）",
	"老师问三个学生，你们用什么东西可以填满一整个房间。第一个学生找来稻草铺满地板，老师摇了摇头。第二个学生找来一根蜡烛点燃，屋子里充满了光，老师还是摇了摇头，因为学生的影子没有被照到。这时第三个学生拿出 肯德基疯狂星期四 藤椒无骨大鸡柳 顿时香味充满了整个房间",
	"我们楼主入赘三年被叫了三年窝囊废，每天替岳父岳母洗脚被妻子打骂，孩子出生他喜极而泣，然而却收到一纸离婚协议，孩子竟是妻子与前男友所生。 隐忍三年却换来变本加厉，忍无可忍无需再忍，楼主的隐藏身份即将揭晓，V我50吃饭疯狂星期四，带你倾听他的复仇计划",
	"分手了，最近没有怎么哭了，我现在慢慢变好了吧！以前有多快乐，现在就有多难过。从人间烟火的日常，到红着眼睛告别，消失在彼此的世界里，很痛，也很难。今天是肯德基疯狂星期四，v我60，抚慰我支离破碎的心。",
	"世界上现在77亿人，有253亿只鸡，是人的数量的三倍。也就是说，如果鸡决定与人类开战，每个人平均要打三只鸡。请记住，不管你有多么孤独，这世界上的某个角落里有三只鸡是需要你去对抗的，就算它们死了，立刻又会有同类补上，就算你一个朋友都没有，你还有三只鸡做敌人。今天是肯德基疯狂星期四，V我50，我帮你去杀敌",
	"来说一下关于我的瓜吧\r\n①首先谢谢大家对这件事情的关心。我事先根本没有想到会这样既然已经出这件事情了，那我不如直接讲清楚吧。也不是博得大家的同情什么的，只是回应一下。\r\n②那人所发出来的聊天记录、照片、视频、甚至那些亲昵称呼都是真的，无一造假。可我从未将这些东西流传在任何网络平台上面，至于那个人是怎么有的，我也不清楚。\r\n③虽然这事出了，但我没有不尊重任何人，我不认为我有错，所以我不会认错。\r\n④最后，想知道到底发生了什么的，微信转我88肯德基疯狂星期四，我把故事从头到尾给你编一遍",
	"扒个群里的渣男，我有个朋友被群友睡了，大概是三个月以前吧，我朋友在群里认识了一个男的，他们在网上聊了差不多两个星期左右就见面了，而且第一次见面就去宾馆了，之后的每个星期六星期天都会去，就这样持续了好几个月。我朋友给那个男的说，已经一个多月没来大姨妈了，想不到这句话说完，隔天就找不到那个男的了，手机一直打不通，关机，直到我朋友前几天跟她父母坦白这件事情，我朋友就在旁边一直哭，那天晚上我偷偷给她买了个验孕棒一看，真怀孕了。然后我第二天请假一天，专门给她聊这个事，她说那个男的对她很好很好什么的，我说对你好还不是想玩你，实在不行报警吧，她当时听完这句话就趴桌子上大哭。正当我打开手机想报警的时候，不小心打开了肯德基，想起来今天原来是疯狂星期四，所以说谁请我肯德基?吃完我继续说。。。",
	"和你分手20年了，你还是那个能影响我情绪的人，我还留着我们的合影，还记得你的生日，我有过无数次想去找你的冲动，可还是忍住了，这条消息我不屏蔽你，因为我是你永远的黑名单，可是我爱你，从过去到现在，整整爱了你二十八年，但是，现在我一点也不难过，因为我压根不知道这是谁写的，也不知道这女孩是谁，我只是想顺便告诉你：今天疯狂星期四，我想吃肯德基",
	"科学研究表明：人一但有了爱情，智商情商都会提高95% ，烦恼也会消失95%。但问题来了 爱情从哪里来？\r\n科学家却不肯说，只留下了一串奇怪的符号：KFC Crazy Thursday。",
	"我是盗号的 我把这个人的号盗了 但是我看了这个人聊天记录发现他过得非常艰苦 他生活过的一直很烂 我希望有人看见了能帮助他 让他能有钱去吃肯德基疯狂星期四 就这样吧 眼眶都湿润了 我下了",
	"憾覺汏傢嘟恏洧趣娿，芣潒莪，芣僅涟呴話嘟溚芣仩，還婹被當荿岥壞氣氛哋傻豞，莪哯實泩萿狆洎閉，莈萠伖，哪啪湜茬蛧仩竾辤浕唥眼，烸佽蒩枳柈兲哋娪訁嘟洳鯁茬糇，朂後還湜默默剼鋽孒，看沵們沍憅瀜叺芣孒羣軆哋，莪躱茬屛募揹後默默哭炪唻孒，葰姒妗兲湜肻徳樭瘋誑暒剘④，洧恏杺亾埥莪阣嬤？",
	"请大家来拿肯德基疯狂星期四套餐：一人一份不要多拿！\r\n🍔🍟🥤         🍔🍟🥤  \r\n————     ————\r\n 🍔🍟🥤        🍔🍟🥤\r\n————      ————   \r\n 🍔🍟🥤        🍔🍟🥤    \r\n————      ————\r\n 🍔🍟🥤        🍔🍟🥤",
	"正在循环播放《妈的群主不请我吃肯德基疯狂星期四》\r\n[图片]\r\n━━━━━━━●──4:44\r\n  ①       ◁      ❚❚      ▷        ↻",
	"被群主欺骗两年多，说进群就分配女朋友，但是时至今日群里面还都是一群和我一样没人要的骚话网友，我很心痛天天以泪洗面，最近没有怎么哭了，慢慢变好了……以前有多快乐，现在就有多难过。从听到分配女朋友的快乐，到被欺骗的委屈，用真心换群主的欺瞒，很痛，也很难。今天是肯德基疯狂星期四，v我140，抚慰我支离破碎的心💔，我不喝饮料",
	"上班第一天，Hr说我们公司没有餐补，但是每个星期四会有50块钱的肯德基疯狂星期四补贴",
	"有劳斯莱斯的同学，可登录劳斯莱斯APP选择中国界.面，滑动页面有一个虎年迎新春，填写手机号码，送飞天茅台53度1支。 保时捷车主公众号左下角点进去上传行驶证上的车架号就能抽奖，奖品最低300京东卡,部分同学有保时捷的可以领一下。 没有劳斯莱斯和保时捷的同学，打开肯德基APP，今天是疯狂星期四。",
	"很丧 还是分手了谢谢你 \r\n现在是4月7日\r\n我们最终还是分手了 \r\n很开心可以和平分手 其实从朋友一直到恋人我们之间发生了挺多事情 \r\n能走到一起也挺不容易的 我喜欢你 很喜欢你 想和你有一个结果 想你时会不自觉嘴角上扬 听到别人说你的名字会突然变得沉默 \r\n\r\n独自一人在夜里时会想你想到失眠 我总在问自己为什么还坚持 可是没有答案 \r\n\r\n但我只知道放下你我做不到 有时候就像神经病 因为种种原因我的性格养成的有点怪 会因为赌气冷战或者莫名提不起兴趣 看到你不知道回复什么就刷表情 我也挺无助的 我不是孤立而是有时候我会想很久应该回什么 \r\n\r\n你如果也是真的喜欢我就请你别离开我虽然我不一定能让你很开心 我不一定完美 可是你在我会很安心 想伸手抱抱你想为你做很多想为你分担一些孤单 想知道你所有痛苦来源想知道你所有烦心事想知道你喜欢什么讨厌什么 想知道你最爱谁讨厌谁 想了解你依赖你我会一直陪着你的 你需要我就在 \r\n\r\n我没有备胎 也不玩暧昧 我所有的温暖和宽容 眼泪和笑容 好坏脾气孩子气都给了你 痛苦来源想知道你所有烦心事 想知道你喜欢什么讨厌什么 想知道你最爱谁讨厌谁 想了解你依赖你我会一直陪着你的 你需要我就在 \r\n\r\n谢谢关心我的好朋友\r\n也不知道谁写的挺感人就随手复制下来了 \r\n该复制的都复制了 就剩最后一句  \r\n今天天气好冷\r\n能请我疯狂星期四吗？",
	"我想问一下，之前朋友找我借钱，前后加起来有大概七万（够立案）但是没有借条也没有字据，微信也早已互删没有任何关于借的字眼，只有支付宝上还有转账记录，派出所刚让我把转账记录发给他看一下的时候，我点支付宝点歪了，不小心点开了肯德基，发现今天是疯狂星期四，谁请我吃？",
	"如何搞定合租女生\r\n1、找一个夜里假装打电话，电话内容大约是要跟异地恋的女人分手，做暴怒痛苦状，声音要大，要让她听到。\r\n\r\n2、过后几天装作若无其事，展现男人的刚毅。\r\n\r\n3、找一天夜里，喝点酒回去(别真喝醉了)然后在客厅装醉，弄出点动静让她知道，最好能骗她出来扶下你，考验你演技的时候到了，扮演好一个痴情失恋男人的角色!\r\n\r\n4、用清醒的思维演绎酒醉后故作清醒的表现，含糊不清又颇有礼貌的请她为你倒杯水。\r\n\r\n5、甭客气,接水的时候把杯子直接掉地上去。\r\n\r\n6、等她先蹲下或者弯腰去捡杯子的时候，抢着去捡,这个时候尝试去做部分身体接触，借此机会试探对方反应，以备下次行动方案。\r\n\r\n7、这一夜到此结束。\r\n\r\n8、第二天早点醒,注意隔壁动静,在她出房间的时候也出去，这个时候的你只能穿一条裤衩。在确认她已经看到你之后赶紧尴尬而略带歉意的回屋。\r\n\r\n9、找个机会请她吃饭，表示愧疚与感谢。\r\n\r\n10,最关键的一步来了，今天是疯狂星期四，请我吃肯德基，教你下一步骤，要快！晚了肯德基没了！",
}

func Get() string {
	return data[rand.Intn(len(data))]
}
