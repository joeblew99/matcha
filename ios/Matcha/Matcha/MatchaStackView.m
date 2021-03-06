#import "MatchaStackView.h"
#import "MatchaView.h"
#import "MatchaProtobuf.h"
#import "MatchaViewController.h"
#import <objc/runtime.h>

#define VIEW_ID_KEY @"matchaViewId"

@interface UIViewController (MatchaStackScreen)
- (void)matcha_setViewId:(int64_t)value;
- (int64_t)matcha_viewId;
@end

@implementation UIViewController (MatchaStackScreen)

- (void)matcha_setViewId:(int64_t)value {
    @synchronized (self) {
        objc_setAssociatedObject(self, VIEW_ID_KEY, @(value), OBJC_ASSOCIATION_RETAIN);
    }
}

- (int64_t)matcha_viewId {
    @synchronized (self) {
        return ((NSNumber *)objc_getAssociatedObject(self, VIEW_ID_KEY)).longLongValue;
    }
}

@end

@implementation MatchaStackView

+ (void)load {
    MatchaRegisterViewController(@"gomatcha.io/matcha/view/stacknav", ^(MatchaViewNode *node){
        return [[MatchaStackView alloc] initWithViewNode:node];
    });
    MatchaRegisterViewController(@"gomatcha.io/matcha/view/stacknav Bar", ^(MatchaViewNode *node){
        return [[MatchaStackBar alloc] initWithViewNode:node];
    });
}

- (id)initWithViewNode:(MatchaViewNode *)viewNode {
    if ((self = [super init])) {
        self.viewNode = viewNode;
        self.delegate = self;
        MatchaConfigureChildViewController(self);
        self.view.backgroundColor = [UIColor whiteColor];
    }
    return self;
}

- (void)setMatchaChildViewControllers:(NSArray<UIViewController *> *)childVCs {
    MatchaStackScreenPBView *view = (id)[self.node.nativeViewState unpackMessageClass:[MatchaStackScreenPBView class] error:nil];
    
    NSMutableArray *prevIds = [NSMutableArray array];
    for (MatchaStackScreenPBChildView *i in view.childrenArray) {
        [prevIds addObject:@(i.screenId)];
    }
    if ([self.prevIds isEqual:prevIds]) {
        return;
    }
    self.prevIds = prevIds;
    self.navigationBar.barTintColor = view.hasBarColor ? [[UIColor alloc] initWithProtobuf:view.barColor] : nil;
    self.navigationBar.titleTextAttributes = view.hasTitleTextStyle ? [NSAttributedString attributesWithProtobuf:view.titleTextStyle] : nil;
    if (view.hasBackTextStyle) {
        [[UIBarButtonItem appearance] setTitleTextAttributes:[NSAttributedString attributesWithProtobuf:view.backTextStyle] forState:UIControlStateNormal];
    }

    NSMutableArray *viewControllers = [NSMutableArray array];
    for (NSInteger i = 0; i < view.childrenArray.count; i++) {
        MatchaStackScreenPBChildView *childView = view.childrenArray[i];
        MatchaStackBar *bar = (id)childVCs[i * 2];
        UIViewController *vc = childVCs[i * 2 + 1];
        vc.navigationItem.title = bar.titleString;
        vc.navigationItem.hidesBackButton = bar.backButtonHidden;
        vc.navigationItem.titleView = bar.titleView;
        vc.navigationItem.rightBarButtonItems = bar.rightViews;
        vc.navigationItem.leftBarButtonItems = bar.leftViews;
        vc.navigationItem.leftItemsSupplementBackButton = true;
        if (bar.customBackButtonTitle) {
            vc.navigationItem.backBarButtonItem = [[UIBarButtonItem alloc] initWithTitle:bar.backButtonTitle style:UIBarButtonItemStylePlain target:nil action:nil];
        }
        [vc matcha_setViewId:childView.screenId];
        [viewControllers addObject:vc];
    }
    
    if (self.viewControllers.count == viewControllers.count) {
        [self setViewControllers:viewControllers animated:NO];
    } else {
        [self setViewControllers:viewControllers animated:YES];
    }
    self.prev = viewControllers;
}

//- (void)navigationController:(UINavigationController *)navigationController willShowViewController:(UIViewController *)viewController animated:(BOOL)animated {
//    NSLog(@"willShow");
//}

- (void)navigationController:(UINavigationController *)navigationController didShowViewController:(UIViewController *)viewController animated:(BOOL)animated {
    [self update];
}

- (void)update {
    NSMutableArray *prevIds = [NSMutableArray array];
    for (UIViewController *i in self.childViewControllers) {
        [prevIds addObject:@(i.matcha_viewId)];
    }
    if ([self.prevIds isEqual:prevIds]) {
        return;
    }
    self.prevIds = prevIds;
    
    GPBInt64Array *array = [[GPBInt64Array alloc] init];
    for (NSNumber *i in prevIds) {
        [array addValue:i.longLongValue];
    }
    MatchaStackScreenPBStackEvent *event = [[MatchaStackScreenPBStackEvent alloc] init];
    event.idArray = array;
    
    MatchaGoValue *value = [[MatchaGoValue alloc] initWithData:event.data];
    [self.viewNode.rootVC call:@"OnChange" viewId:self.node.identifier.longLongValue args:@[value]];
}

- (void)setMatchaChildLayout:(GPBInt64ObjectDictionary *)layoutPaintNodes {
    // no-op
}

@end

@implementation MatchaStackBar

- (id)initWithViewNode:(MatchaViewNode *)viewNode {
    if ((self = [super init])) {
        self.viewNode = viewNode;
    }
    return self;
}

- (void)setMatchaChildViewControllers:(NSArray<UIViewController *> *)childVCs {
    MatchaStackScreenPBBar *bar = (id)[self.node.nativeViewState unpackMessageClass:[MatchaStackScreenPBBar class] error:nil];
    NSInteger idx = 0;
    
    self.titleString = bar.title;
    self.backButtonHidden = bar.backButtonHidden;
    self.backButtonTitle = bar.backButtonTitle;
    self.customBackButtonTitle = bar.customBackButtonTitle;
    if (bar.hasTitleView) {
        self.titleView = childVCs[idx].view;
        idx += 1;
    } else {
        self.titleView = nil;
    }

    NSMutableArray *rightViews = [NSMutableArray array];
    for (NSInteger i = 0; i < bar.rightViewCount; i++) {
        UIView *rightView = childVCs[idx].view;
        UIBarButtonItem *item = [[UIBarButtonItem alloc] initWithCustomView:rightView];
        [rightViews addObject:item];
        idx += 1;
    }
    self.rightViews = rightViews;
    
    NSMutableArray *leftViews = [NSMutableArray array];
    for (NSInteger i = 0; i < bar.leftViewCount; i++) {
        UIView *leftView = childVCs[idx].view;
        UIBarButtonItem *item = [[UIBarButtonItem alloc] initWithCustomView:leftView];
        [leftViews addObject:item];
        idx +=1;
    }
    self.leftViews = leftViews;
}

- (void)setMatchaChildLayout:(NSArray<MatchaViewPBLayoutPaintNode *> *)layoutPaintNodes {
    NSInteger idx = 0;
    if (self.titleView) {
        CGRect f = self.titleView.frame;
        f.size = ((MatchaViewPBLayoutPaintNode *)layoutPaintNodes[idx]).frame.size;
        self.titleView.frame = f;
        idx += 1;
    }
    for (NSInteger i = 0; i < self.rightViews.count; i++) {
        UIBarButtonItem *rightView = self.rightViews[i];
        CGRect f = rightView.customView.frame;
        f.size =((MatchaViewPBLayoutPaintNode *)layoutPaintNodes[idx]).frame.size;
        rightView.customView.frame = f;
        idx += 1;
    }
    for (NSInteger i = 0; i < self.leftViews.count; i++) {
        UIBarButtonItem *leftView = self.leftViews[i];
        CGRect f = leftView.customView.frame;
        f.size =((MatchaViewPBLayoutPaintNode *)layoutPaintNodes[idx]).frame.size;
        leftView.customView.frame = f;
        idx += 1;
    }
}

@end
