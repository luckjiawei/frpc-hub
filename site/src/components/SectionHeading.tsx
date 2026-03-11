import { Icon } from "@iconify/react";
import { Box, Flex, Text } from "@radix-ui/themes";

interface SectionHeadingProps {
  id: string;
  title: string;
  icon: string;
}

export function SectionHeading({ id, title, icon }: SectionHeadingProps) {
  return (
    <div id={id} style={{ scrollMarginTop: 24 }}>
      <Flex align="center" gap="2" mb="4">
        <Box
          style={{
            background: "linear-gradient(135deg, var(--accent-9), var(--accent-10))",
            borderRadius: 8,
            padding: "6px",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            flexShrink: 0,
          }}
        >
          <Icon icon={icon} width="16" height="16" color="white" />
        </Box>
        <Text size="4" weight="bold">
          {title}
        </Text>
      </Flex>
    </div>
  );
}
