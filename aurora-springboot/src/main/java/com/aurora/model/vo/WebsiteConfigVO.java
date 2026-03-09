package com.aurora.model.vo;

import io.swagger.v3.oas.annotations.media.Schema;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@AllArgsConstructor
@NoArgsConstructor
@Builder
@Schema(description = "网站配置")
public class WebsiteConfigVO {

    @Schema(name = "name", description = "网站名称", required = true)
    private String name;

    @Schema(name = "nickName", description = "网站作者昵称", required = true)
    private String englishName;

    @Schema(name = "author", description = "网站作者", required = true)
    private String author;

    @Schema(name = "avatar", description = "网站头像", required = true)
    private String authorAvatar;

    @Schema(name = "description", description = "网站作者介绍", required = true)
    private String authorIntro;

    @Schema(name = "logo", description = "网站logo", required = true)
    private String logo;

    @Schema(name = "multiLanguage", description = "多语言", required = true)
    private Integer multiLanguage;

    @Schema(name = "notice", description = "网站公告", required = true)
    private String notice;

    @Schema(name = "websiteCreateTime", description = "网站创建时间", required = true)
    private String websiteCreateTime;

    @Schema(name = "beianNumber", description = "网站备案号", required = true)
    private String beianNumber;

    @Schema(name = "qqLogin", description = "QQ登录", required = true)
    private Integer qqLogin;

    @Schema(name = "github", description = "github", required = true)
    private String github;

    @Schema(name = "gitee", description = "gitee", required = true)
    private String gitee;

    @Schema(name = "qq", description = "qq", required = true)
    private String qq;

    @Schema(name = "weChat", description = "微信", required = true)
    private String weChat;

    @Schema(name = "weibo", description = "微博", required = true)
    private String weibo;

    @Schema(name = "csdn", description = "csdn", required = true)
    private String csdn;

    @Schema(name = "zhihu", description = "zhihu", required = true)
    private String zhihu;

    @Schema(name = "juejin", description = "juejin", required = true)
    private String juejin;

    @Schema(name = "twitter", description = "twitter", required = true)
    private String twitter;

    @Schema(name = "stackoverflow", description = "stackoverflow", required = true)
    private String stackoverflow;

    @Schema(name = "touristAvatar", description = "游客头像", required = true)
    private String touristAvatar;

    @Schema(name = "userAvatar", description = "用户头像", required = true)
    private String userAvatar;

    @Schema(name = "isCommentReview", description = "是否评论审核", required = true)
    private Integer isCommentReview;

    @Schema(name = "isEmailNotice", description = "是否邮箱通知", required = true)
    private Integer isEmailNotice;

    @Schema(name = "isReward", description = "是否打赏", required = true)
    private Integer isReward;

    @Schema(name = "weiXinQRCode", description = "微信二维码", required = true)
    private String weiXinQRCode;

    @Schema(name = "alipayQRCode", description = "支付宝二维码", required = true)
    private String alipayQRCode;

    @Schema(name = "favicon", description = "favicon", required = true)
    private String favicon;

    @Schema(name = "websiteTitle", description = "网页标题", required = true)
    private String websiteTitle;

    @Schema(name = "gonganBeianNumber", description = "公安部备案编号", required = true)
    private String gonganBeianNumber;

}
