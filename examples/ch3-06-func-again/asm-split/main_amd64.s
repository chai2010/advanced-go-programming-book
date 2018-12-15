
#define NOSPLIT 4

TEXT ·printnl_nosplit(SB), NOSPLIT, $8
	CALL runtime·printnl(SB)
	RET

TEXT ·printnl(SB), $8
	CALL runtime·printnl(SB)
	RET
